package core

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	hex "encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
	"microsoft.com/azure-spring-discovery/api/logging"
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
)

type AuthType int32

var credentialCache sync.Map

type DiscoveryExecutor interface {
	Prepare() (*Credential, error)
	Server() TargetServer
	ProcessScan() ([]JavaProcess, error)
	GetTotalMemoryInKB() (float64, error)
	GetDefaultMaxHeapSizeInBytes() (float64, error)
	ReadJarFile(location string, walkers ...JarFileWalker) (JarFile, error)
	Finish() error
}

type linuxDiscoveryExecutor struct {
	credentialProvider CredentialProvider
	server             TargetServer
	ctx                context.Context
}

var DefaultJarFileWalkers = []JarFileWalker{
	appConfigWalker,
	loggingConfigWalker,
	manifestWalker,
	certWalker,
	dependencyWalker,
	staticContentWalker,
	pomFileWalker,
}

type JarFileWalker func(name string, f *zip.File, j *jarFile) error

var appConfigWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if isAppConfig(name) {
		content, err := readFileInArchive(f)
		if err != nil {
			return err
		}
		j.applicationConfigurations[strings.ReplaceAll(name, DefaultClasspath, "")] = content
	}
	return nil
}

var loggingConfigWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if isLoggingConfig(name) {
		content, err := readFileInArchive(f)
		if err != nil {
			return err
		}
		j.loggingConfigs[strings.ReplaceAll(name, DefaultClasspath, "")] = content
	}
	return nil
}

var certWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if isCertificate(name) {
		j.certificates = append(j.certificates, strings.ReplaceAll(name, DefaultClasspath, ""))
	}
	return nil
}

var manifestWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if filepath.Base(name) == ManifestFile {
		content, err := readFileInArchive(f)
		if err != nil {
			return err
		}
		j.manifests = parseManifests(content)
	}
	return nil
}

var dependencyWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if filepath.Ext(name) == JarFileExt {
		j.dependencies = append(j.dependencies, strings.ReplaceAll(name, DefaultLibPath, ""))
	}
	return nil
}

var staticContentWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if isStaticContent(name) {
		j.staticFiles = append(j.staticFiles, strings.ReplaceAll(name, DefaultClasspath, ""))
	}
	return nil
}

var pomFileWalker JarFileWalker = func(name string, f *zip.File, j *jarFile) error {
	if isPomFile(name) {
		content, err := readFileInArchive(f)
		if err != nil {
			return err
		}
		j.mvnProject, err = readPom(content)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewLinuxDiscovery(ctx context.Context, targetServer TargetServer, credentialProvider CredentialProvider) DiscoveryExecutor {
	return &linuxDiscoveryExecutor{server: targetServer, credentialProvider: credentialProvider, ctx: ctx}
}

func (l *linuxDiscoveryExecutor) Server() TargetServer {
	return l.server
}

func (l *linuxDiscoveryExecutor) Prepare() (*Credential, error) {
	var err error
	parallelEnabled := config.GetConfigBoolByKey(ParallelConnectConfigKey)
	azureLogger := logging.GetAzureLogger(l.ctx)
	cred, cached := getCredFromCache(l.server.FQDN())
	if cached {
		azureLogger.Debug("credential cache hit", "server", l.server.FQDN(), "credentialId", cred.Id)
		if parallelEnabled {
			_, err = l.server.ParallelConnect(cred)
		} else {
			_, err = l.server.Connect(cred)
		}
	}

	if cached && err == nil {
		return cred, nil
	}

	azureLogger.Info("credential cache miss or login failed, going to get credentials from K8s", "server", l.server.FQDN(), "cached", cached, "parallelEnabled", parallelEnabled, "err", err)
	creds, err := l.credentialProvider.GetCredentials()
	if err != nil {
		return nil, CredentialError{message: fmt.Sprintf("failed to get credentials, %s", err)}
	}
	if parallelEnabled {
		cred, err = l.server.ParallelConnect(creds...)
	} else {
		cred, err = l.server.Connect(creds...)
	}
	if err == nil {
		azureLogger.Debug("going to cache the credential", "server", l.server.FQDN(), "credentialId", cred.Id)
		cacheCredential(l.server.FQDN(), cred)
	}
	return cred, err
}

func (l *linuxDiscoveryExecutor) ProcessScan() ([]JavaProcess, error) {

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
				return nil, DiscoveryError{message: fmt.Sprintf("failed to parse pid, %s", splits[0]), severity: v1alpha1.Error}
			}
			uid, err = strconv.Atoi(splits[1])
			if err != nil {
				return nil, DiscoveryError{message: fmt.Sprintf("failed to parse uid, %s", splits[1]), severity: v1alpha1.Error}
			}
			start := 2
			for _, split := range splits[start:] {
				if strings.HasSuffix(split, JavaCmd) {
					break
				}
				start++
			}
			if start >= len(splits) {
				return nil, DiscoveryError{message: fmt.Sprintf("cannot locate java command in scanned process options, %s", splits), severity: v1alpha1.Error}
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

func (l *linuxDiscoveryExecutor) GetTotalMemoryInKB() (float64, error) {
	output, err := l.server.RunCmd(GetTotalMemoryCmd())
	if err != nil {
		return 0, err
	}
	if len(output) == 0 {
		return 0, DiscoveryError{message: fmt.Sprintf("failed to get total memory, output is empty"), severity: v1alpha1.Warning}
	}

	size, err := strconv.ParseFloat(cleanOutput(output, LinuxNewLineCharacter), 64)
	if err != nil {
		return 0, DiscoveryError{message: fmt.Sprintf("unable to parse total memory, output is %s", output), severity: v1alpha1.Warning}
	}

	return size, nil
}

func (l *linuxDiscoveryExecutor) GetDefaultMaxHeapSizeInBytes() (float64, error) {
	output, err := l.server.RunCmd(LinuxGetDefaultMaxHeapCmd)
	if err != nil {
		return 0, err
	}
	if len(output) == 0 {
		return 0, DiscoveryError{message: fmt.Sprintf("failed to get default MaxHeapSize, output is empty"), severity: v1alpha1.Warning}
	}

	size, err := strconv.ParseFloat(cleanOutput(output, LinuxNewLineCharacter), 64)
	if err != nil {
		return 0, DiscoveryError{message: fmt.Sprintf("failed to parse default MaxHeapSize, output: %s", output), severity: v1alpha1.Warning}
	}

	return size, nil
}

func (l *linuxDiscoveryExecutor) getChecksum(absolutePath string) (string, error) {
	azureLogger := logging.GetAzureLogger(l.ctx)
	output, err := l.Server().RunCmd(GetSha256Cmd(absolutePath))
	if err != nil || len(output) == 0 {
		azureLogger.Info("cannot get sha256 checksum", "absolutePath", absolutePath, "err", err)
		return "", nil
	}
	return cleanOutput(output, LinuxNewLineCharacter), nil
}

func (l *linuxDiscoveryExecutor) ReadJarFile(location string, walkers ...JarFileWalker) (JarFile, error) {
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
		return nil, DiscoveryError{message: fmt.Sprintf("cannot read remote location: %s, %s", location, err.Error()), severity: v1alpha1.Error}
	}

	checksum, _ := l.getChecksum(location)
	if r, ok := srcFile.(io.Reader); ok && len(checksum) == 0 {
		// this step will slow down the overall speed
		h := sha256.New()
		if _, err := io.Copy(h, r); err != nil {
			return nil, DiscoveryError{message: fmt.Sprintf("cannot get checksum for %s, %s", location, err.Error()), severity: v1alpha1.Error}
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

func (l *linuxDiscoveryExecutor) Finish() error {
	return l.Server().Close()
}

func cleanOutput(raw string, newLine string) string {
	return strings.TrimSuffix(strings.TrimSpace(raw), newLine)
}

func sanitizeVersion(version string) string {
	return strings.ReplaceAll(version, "_", "-")
}

func sudo(command string) string {
	return "sudo " + command
}

func getCredFromCache(server string) (*Credential, bool) {
	if v, ok := credentialCache.Load(server); ok {
		return v.(*Credential), true
	}
	return nil, false
}

func cacheCredential(server string, credential *Credential) {
	credentialCache.Store(server, credential)
}
