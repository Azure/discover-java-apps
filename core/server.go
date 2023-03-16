package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"microsoft.com/azure-spring-discovery/api/logging"
)

const (
	ParallelConnectConfigKey = "config.springbootserver.connect.parallel"
)

type TargetServer interface {
	FQDN() string
	Connect(creds ...*Credential) (*Credential, error)
	ParallelConnect(creds ...*Credential) (*Credential, error)
	Close() error
	Read(remoteLocation string) (io.ReaderAt, os.FileInfo, error)
	RunCmd(cmd string) (string, error)
}

type TargetServerFactory interface {
	Create(ctx context.Context, host string, port int) TargetServer
}

type linuxServerFactory struct {
}

func DefaultServerFactory() TargetServerFactory {
	return &linuxServerFactory{}
}

func (f *linuxServerFactory) Create(ctx context.Context, host string, port int) TargetServer {
	return &linuxServer{
		server: host,
		port:   port,
		ctx:    ctx,
	}
}

type linuxServer struct {
	server string
	port   int
	client *ssh.Client
	ctx    context.Context
}

func (s *linuxServer) RunCmd(cmd string) (string, error) {
	if s.client == nil {
		return "", ConnectionError{error: fmt.Errorf("server %s is not connected", s.server)}
	}
	var session *ssh.Session
	var err error

	session, err = s.client.NewSession()
	if err != nil {
		return "", ConnectionError{error: err, message: fmt.Sprintf("failed to create new session, server: %s", s.server)}
	}
	defer session.Close()
	var b bytes.Buffer
	var e bytes.Buffer
	session.Stdout = &b
	session.Stderr = &e
	err = session.Run(cmd)
	azureLogger := logging.GetAzureLogger(s.ctx)
	if strings.Contains(e.String(), "Permission denied") {
		err = PermissionDenied{error: err, message: e.String()}
		azureLogger.Error(err, "Running cmd on server permission denied", "cmd", cmd, "server", s.server, "output", e.String())
		return "", err
	}
	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			if exitError.ExitStatus() == 1 && strings.Contains(cmd, "grep") {
				// for a grep command, exit code = 1 means no lines were returned
				// https://linuxcommand.org/lc3_man_pages/grep1.html
				return "", nil
			}
		}
		azureLogger.Debug("Running cmd on server failed", "cmd", cmd, "server", s.server, "err", err, "output", e.String())
		return "", toSshError(err, &e)
	}

	return b.String(), nil
}

func (s *linuxServer) Connect(creds ...*Credential) (*Credential, error) {
	length := len(creds)
	if length == 0 {
		return nil, CredentialError{message: "credentials are empty"}
	}

	var connError error
	azureLogger := logging.GetAzureLogger(s.ctx)
	for _, cred := range creds {
		client, err := s.doConnect(cred, time.Second*3)
		if err != nil {
			if strings.Contains(err.Error(), "ssh: unable to authenticate") {
				// do nothing
			} else {
				azureLogger.Debug("error to connect to server with credential", "server", s.server, "cred", cred.FriendlyName, "err", err.Error())
				connError = err
				break
			}
		} else {
			s.client = client
			return cred, nil
		}
	}

	if connError != nil {
		if strings.Contains(connError.Error(), "connection reset by peer") {
			return nil, RetryableError{connError}
		}
		return nil, ConnectionError{error: connError, message: fmt.Sprintf("connect to server failed, server: %s", s.server)}
	}

	return nil, CredentialError{message: fmt.Sprintf("tried all credentials, but still cannot connect to server: %s ", s.server)}
}

func (s *linuxServer) ParallelConnect(creds ...*Credential) (*Credential, error) {
	length := len(creds)
	if length == 0 {
		return nil, CredentialError{message: "credentials are empty"}
	}

	ch := make(chan loginResult)
	defer close(ch)
	timeoutInSeconds := config.GetConfigIntByKey("config.springbootserver.connect.timeoutSeconds")
	azureLogger := logging.GetAzureLogger(s.ctx)
	for _, credential := range creds {
		go func(cred *Credential, ch chan<- loginResult) {
			client, errInner := s.doConnect(cred, time.Second*3)
			defer func() {
				if r := recover(); r != nil {
					if client != nil {
						azureLogger.Debug("recovered from panic, close connection id", "id", client.SessionID(), "r", r)
						_ = client.Close()
					}
				}
			}()
			ch <- loginResult{
				cred:   cred,
				client: client,
				err:    errInner,
			}
		}(credential, ch)
	}
	t := time.NewTimer(time.Second * time.Duration(timeoutInSeconds))
	finished := 0
	completed := false
	var timeout = false
	var cred *Credential
	var connError error
	for {
		if completed || finished == length {
			break
		}
		select {
		case r := <-ch:
			finished++
			if r.client != nil {
				s.client = r.client
				cred = r.cred
				completed = true
			} else if r.err != nil {
				if strings.Contains(r.err.Error(), "ssh: unable to authenticate") {
					// do nothing
				} else {
					azureLogger.Debug("error to connect to server", "server", s.server, "err", r.err)
					connError = r.err
				}
			}
		case <-t.C:
			timeout = true
			completed = true
		}
	}

	var err error
	if cred == nil {
		if timeout {
			err = RetryableError{fmt.Errorf("connect to server %s, timed out after %d seconds", s.server, 30)}
		} else {
			if connError != nil {
				err = ConnectionError{error: connError, message: fmt.Sprintf("connect to server failed, server: %s", s.server)}
			} else {
				err = CredentialError{message: fmt.Sprintf("tried all credentials, but still cannot connect to %s ", s.server)}
			}
		}

	}
	return cred, err
}

func (s *linuxServer) Close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *linuxServer) Read(location string) (io.ReaderAt, os.FileInfo, error) {
	_, _, err := s.client.SendRequest("keepalive", false, nil)
	if err != nil {
		return nil, nil, err
	}
	client, err := sftp.NewClient(s.client)
	if err != nil {
		return nil, nil, ConnectionError{error: err, message: fmt.Sprintf("create sftp client failed, server: %s, location: %s", s.server, location)}
	}

	// Read the source file
	srcFile, err := client.Open(location)
	if err != nil {
		return nil, nil, ConnectionError{error: err, message: fmt.Sprintf("read jar file over sftp failed, server: %s, location: %s", s.server, location)}
	}

	stat, err := srcFile.Stat()
	if err != nil {
		return nil, nil, ConnectionError{error: err, message: fmt.Sprintf("stat jar file failed: server: %s, location: %s", s.server, location)}
	}

	return srcFile, stat, nil
}

func (s *linuxServer) FQDN() string {
	return s.server
}

func toSshError(err error, output *bytes.Buffer) error {
	if err == nil {
		return nil
	}
	return &SshError{error: err, message: output.String()}
}

func (s *linuxServer) doConnect(cred *Credential, timeout time.Duration) (*ssh.Client, error) {
	azureLogger := logging.GetAzureLogger(s.ctx)
	var auth []ssh.AuthMethod
	auth = append(auth, ssh.Password(cred.Password))
	pemBytes := []byte(cred.Password)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		//s.log.Info("failed to parse public key, using password login only", "err", err, "server", s.server)
	} else {
		auth = append(auth, ssh.PublicKeys(signer))
	}
	cfg := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		BannerCallback: func(message string) error {
			azureLogger.Info(message)
			return nil
		},
		Timeout: timeout, // connect timeout
	}
	// connect ot ssh server
	connectString := fmt.Sprintf("%s:%d", s.server, s.port)
	return ssh.Dial("tcp", connectString, cfg)
}

type loginResult struct {
	cred   *Credential
	client *ssh.Client
	err    error
}
