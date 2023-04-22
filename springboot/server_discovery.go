package springboot

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"strconv"
	"strings"
)

type AuthType int32

const (
	ManifestFile = "MANIFEST.MF"
	JarFileExt   = ".jar"
)

type linuxServerDiscovery struct {
	credentialProvider CredentialProvider
	server             ServerConnector
	ctx                context.Context
	cfg                YamlConfig
}

func NewLinuxServerDiscovery(
	ctx context.Context,
	serverConnector ServerConnector,
	credentialProvider CredentialProvider,
	cfg YamlConfig) ServerDiscovery {
	return &linuxServerDiscovery{
		ctx:                ctx,
		cfg:                cfg,
		server:             serverConnector,
		credentialProvider: credentialProvider,
	}
}

func (l *linuxServerDiscovery) Server() ServerConnector {
	return l.server
}

func (l *linuxServerDiscovery) Prepare() (*Credential, error) {
	creds, err := l.credentialProvider.GetCredentials()
	if err != nil {
		return nil, CredentialError{error: err, message: "failed to get credentials"}
	}
	return l.connect(creds...)
}

func (l *linuxServerDiscovery) ProcessScan() ([]JavaProcess, error) {

	var output string
	var err error
	output, err = runWithSudo(l.Server(), GetProcessScanCmd())
	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			if exitError.ExitStatus() == 1 {
				// when ps command return empty processes, linux will return the exit status 1
				// we just ignore this kind of error
			}
		} else {
			return nil, err
		}

	}
	var processes []JavaProcess

	if len(output) == 0 {
		return processes, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		var process JavaProcess
		process, err = func(line string) (JavaProcess, error) {
			splits := strings.Fields(strings.TrimSpace(line))
			var pid int
			var uid int
			pid, err = strconv.Atoi(splits[0])
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse pid from process")
			}
			uid, err = strconv.Atoi(splits[1])
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse uid from process")
			}
			start := 2
			for _, split := range splits[start:] {
				if strings.HasSuffix(split, JavaCmd) {
					break
				}
				start++
			}
			if start >= len(splits) {
				return nil, errors.New("cannot locate java command in scanned process options")
			}
			return &javaProcess{
				pid:      pid,
				uid:      uid,
				javaCmd:  splits[start],
				options:  splits[start+1:],
				executor: l,
			}, nil
		}(line)
		if err != nil {
			return nil, err
		}
		processes = append(processes, process)
	}
	return processes, nil
}

func (l *linuxServerDiscovery) GetTotalMemory() (int64, error) {
	output, err := runWithSudo(l.server, GetTotalMemoryCmd())
	if err != nil {
		return 0, err
	}
	if len(output) == 0 {
		return 0, errors.New("failed to get total memory, output is empty")
	}

	size, err := strconv.ParseFloat(CleanOutput(output), 64)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("unable to parse total memory, output is %s", output))
	}

	return int64(size * KiB), nil
}

func (l *linuxServerDiscovery) getChecksum(absolutePath string) (string, error) {
	azureLogger := GetAzureLogger(l.ctx)
	output, err := runWithSudo(l.server, GetSha256Cmd(absolutePath))
	if err != nil || len(output) == 0 {
		azureLogger.Info("cannot get sha256 checksum", "absolutePath", absolutePath, "err", err)
		return "", nil
	}
	return CleanOutput(output), nil
}

func (l *linuxServerDiscovery) GetOsName() (string, error) {
	azureLogger := GetAzureLogger(l.ctx)
	var tryOsRelease tryFunc[ServerConnector, string] = func(in ServerConnector) (string, bool) {
		output, err := runWithSudo(in, GetOsName())
		if err != nil {
			azureLogger.Warning(err, "cannot get os name", "output", output)
		}
		return output, len(output) > 0
	}
	var tryCentOsRelease tryFunc[ServerConnector, string] = func(in ServerConnector) (string, bool) {
		output, err := runWithSudo(in, GetCentOsName())
		if err != nil {
			azureLogger.Warning(err, "cannot get cent os name", "output", output)
		}
		return output, len(output) > 0
	}

	output, found := tryFuncs[ServerConnector, string]{tryOsRelease, tryCentOsRelease}.try(l.server)
	if found {
		return CleanOutput(output), nil
	}

	return "", nil
}

func (l *linuxServerDiscovery) GetOsVersion() (string, error) {
	azureLogger := GetAzureLogger(l.ctx)
	var tryOsRelease tryFunc[ServerConnector, string] = func(in ServerConnector) (string, bool) {
		output, err := runWithSudo(in, GetOsVersion())
		if err != nil {
			azureLogger.Debug("cannot get os version", "err", err, "output", output)
		}
		return output, len(output) > 0
	}
	var tryCentOsRelease tryFunc[ServerConnector, string] = func(in ServerConnector) (string, bool) {
		output, err := runWithSudo(in, GetCentOsVersion())
		if err != nil {
			azureLogger.Debug("cannot get cent os version", "err", err, "output", output)
		}
		return output, len(output) > 0
	}

	output, found := tryFuncs[ServerConnector, string]{tryOsRelease, tryCentOsRelease}.try(l.server)
	if found {
		return CleanOutput(output), nil
	}

	return "", nil
}

func (l *linuxServerDiscovery) ReadJarFile(location string, walkers ...JarFileWalker) (JarFile, error) {
	srcFile, fileInfo, err := l.server.Read(location)
	if err != nil {
		return nil, err
	}
	if closer, ok := srcFile.(io.Closer); ok {
		defer closer.Close()
	}

	var reader *zip.Reader
	reader, err = zip.NewReader(srcFile, fileInfo.Size())

	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("cannot read remote location: %s, %s", location, err.Error()))
	}

	checksum, _ := l.getChecksum(location)
	if r, ok := srcFile.(io.Reader); ok && len(checksum) == 0 {
		// this step will slow down the overall speed
		h := sha256.New()
		if _, err := io.Copy(h, r); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot get checksum for %s, %s", location, err.Error()))
		}

		checksum = hex.EncodeToString(h.Sum(nil))
	}

	j := &jarFile{
		checksum:                  checksum,
		remoteLocation:            location,
		applicationConfigurations: make(map[string]string),
		loggingConfigs:            make(map[string]string),
		manifests:                 make(map[string]string),
		lastModifiedTime:          fileInfo.ModTime(),
		size:                      fileInfo.Size(),
	}
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		for _, walker := range walkers {
			err = walker(f.Name, f, j)
			if err != nil {
				return nil, err
			}
		}
	}

	return j, nil
}

func (l *linuxServerDiscovery) Finish() error {
	return l.Server().Close()
}

func sudo(command string) string {
	return "sudo " + command
}

func (l *linuxServerDiscovery) connect(creds ...*Credential) (*Credential, error) {
	azureLogger := GetAzureLogger(l.ctx)
	length := len(creds)
	if length == 0 {
		return nil, CredentialError{error: fmt.Errorf("credentials are empty"), message: ""}
	}

	s := FromSlice[*Credential](l.ctx, creds)
	if l.cfg.Server.Connect.Parallel {
		s = s.Parallel(l.cfg.Server.Connect.Parallelism)
	}

	cred, err :=
		s.Map(func(cred *Credential) (*Credential, error) {
			err := l.server.Connect(cred.Username, cred.Password)
			if err != nil {
				if !isAuthFailure(err) {
					return nil, err
				}
			}
			return cred, nil
		}).Filter(func(t any) bool {
			return t != nil
		}).First()

	if cred != nil {
		return cred.(*Credential), nil
	}

	if err != nil {
		azureLogger.Warning(err, "error to connect to server with credential", "server", l.server.FQDN())
		return nil, ConnectionError{error: err, message: fmt.Sprintf("failed to connect to server: %s", l.server.FQDN())}
	}

	return nil, CredentialError{
		error:   errors.New(fmt.Sprintf("cannot connect to server %s", l.server)),
		message: fmt.Sprintf("tried all credentials, but still cannot connect to server: %s", l.server),
	}
}

func isAuthFailure(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "ssh: unable to authenticate")
}
