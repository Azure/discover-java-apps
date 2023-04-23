package springboot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type linuxServerFactory struct {
	opts []SshOption
}

type SshOption func(s *linuxServer)

func DefaultServerConnectorFactory(opts ...SshOption) ServerConnectorFactory {
	return &linuxServerFactory{opts: opts}
}

func (f *linuxServerFactory) Create(ctx context.Context, host string, port int) ServerConnector {
	s := &linuxServer{
		server: host,
		port:   port,
		ctx:    ctx,
	}

	for _, opt := range f.opts {
		opt(s)
	}
	return s
}

type linuxServer struct {
	client   *ssh.Client
	username string
	cb       ssh.HostKeyCallback
	keyAlgos []string
	timeout  time.Duration
	server   string
	port     int
	ctx      context.Context
	mux      sync.Mutex
}

func (s *linuxServer) RunCmd(cmd string) (string, error) {
	if s.client == nil {
		return "", ConnectionError{error: fmt.Errorf("server %s is not connected", s.server), message: "ssh client is nil"}
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
	output := b.String()
	azureLogger := GetAzureLogger(s.ctx)
	azureLogger.logr.V(1).Info("Running cmd on server", "cmd", cmd, "server", s.server)
	if strings.Contains(e.String(), "Permission denied") {
		err = PermissionDenied{error: fmt.Errorf("run cmd by user %s permission denied", s.username), message: CleanOutput(e.String())}
		azureLogger.Warning(err, "Running cmd permission denied", "cmd", cmd, "server", s.server, "output", CleanOutput(e.String()), "username", s.username)
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
		azureLogger.Warning(err, "Running cmd on server failed", "cmd", cmd, "server", s.server, "output", e.String())
		return "", toSshError(err, &e)
	}

	return output, nil
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
	srcFile, err := client.OpenFile(location, os.O_RDONLY)
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

func (s *linuxServer) Connect(username, password string) error {
	azureLogger := GetAzureLogger(s.ctx)
	var auth []ssh.AuthMethod
	auth = append(auth, ssh.Password(password))
	pemBytes := []byte(password)
	signer, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		//s.log.Info("failed to parse public key, using password login only", "err", err, "server", s.server)
	} else {
		auth = append(auth, ssh.PublicKeys(signer))
	}
	cfg := &ssh.ClientConfig{
		User:            username,
		Auth:            auth,
		HostKeyCallback: s.cb,
		BannerCallback: func(message string) error {
			azureLogger.Info(message)
			return nil
		},
		Timeout: s.timeout, // connect timeout
	}

	cfg.SetDefaults()
	if s.keyAlgos != nil {
		cfg.KeyExchanges = append(cfg.KeyExchanges, s.keyAlgos...)
	}
	cfg.MACs = append(cfg.MACs, "ssh-dss")

	// connect ot ssh server
	connectString := fmt.Sprintf("%s:%d", s.server, s.port)
	client, err := ssh.Dial("tcp", connectString, cfg)
	if err != nil {
		return ConnectionError{error: err, message: fmt.Sprintf("error dial %s", connectString)}
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.client = client
	s.username = username

	return nil
}

func (s *linuxServer) Username() string {
	return s.username
}

func toSshError(err error, output *bytes.Buffer) error {
	if err == nil {
		return nil
	}
	return &SshError{error: err, message: output.String()}
}

func WithHostKeyCallback(callback ssh.HostKeyCallback) SshOption {
	return func(s *linuxServer) {
		s.cb = callback
	}
}

func WithClient(client *ssh.Client) SshOption {
	return func(s *linuxServer) {
		s.client = client
	}
}

func WithConnectionTimeout(timeout time.Duration) SshOption {
	return func(s *linuxServer) {
		s.timeout = timeout
	}
}

func WithKeyAlgorithms(algos []string) SshOption {
	return func(s *linuxServer) {
		s.keyAlgos = algos
	}
}

func (s *linuxServer) String() string {
	return s.server
}
