package springboot

import (
	"context"
	"github.com/creekorful/mvnparser"
	"io"
	"os"
	"time"
)

type AppType string

type AppTypes []AppType

const (
	SpringBootFatJar   AppType = "SpringBootFatJar"
	SpringBootThinJar  AppType = "SpringBootThinJar"
	SpringBootExploded AppType = "SpringBootExploded"
	ExecutableJar      AppType = "ExecutableJar"
	Unknown            AppType = "Unknown"
)

var SpringBootAppTypes = AppTypes{
	SpringBootFatJar,
	SpringBootThinJar,
	SpringBootExploded,
}

func (types AppTypes) Contains(appType AppType) bool {
	for _, t := range types {
		return t == appType
	}
	return false
}

type ServerConnectionInfo struct {
	Server string
	Port   int
}

type Artifact struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Runtime struct {
	Server            string   `json:"server"`
	Uid               int      `json:"uid"`
	Pid               int      `json:"pid"`
	RuntimeJdkVersion string   `json:"runtimeJdkVersion"`
	AppPort           int      `json:"appPort"`
	JavaCmd           string   `json:"javaCmd"`
	Environments      []string `json:"environments"`
	JvmOptions        []string `json:"jvmOptions"`
	JvmMemory         int64    `json:"jvmMemory"`
	OsName            string   `json:"osName"`
	OsVersion         string   `json:"osVersion"`
	BindingPorts      []int    `json:"bindingPorts"`
}

type SpringBootApp struct {
	AppName                   string            `json:"appName"`
	AppType                   AppType           `json:"appType"`
	ApplicationConfigurations map[string]string `json:"applicationConfigurations"`
	Artifact                  *Artifact         `json:"artifact"`
	BuildJdkVersion           string            `json:"buildJdkVersion"`
	Checksum                  string            `json:"checksum"`
	Certificates              []string          `json:"certificates"`
	Dependencies              []string          `json:"dependencies"`
	JarFileLocation           string            `json:"jarFileLocation"`
	JarSize                   int64             `json:"jarSize"`
	LoggingConfigurations     map[string]string `json:"loggingConfigurations"`
	LastModifiedTime          time.Time         `json:"lastModifiedTime"`
	LastUpdatedTime           time.Time         `json:"lastUpdatedTime"`
	Runtime                   *Runtime          `json:"runtime"`
	SpringBootVersion         string            `json:"springBootVersion"`
	StaticContentLocations    []string          `json:"staticContentLocations"`
}

type Credential struct {
	Id             string `json:"Id,omitempty"`
	FriendlyName   string `json:"FriendlyName,omitempty"`
	Username       string `json:"UserName,omitempty"`
	Password       string `json:"Password,omitempty"`
	CredentialType string `json:"CredentialType,omitempty"`
}

type DiscoveryExecutor interface {
	Discover(ctx context.Context, server ServerConnectionInfo, alternativeConnectionInfos ...ServerConnectionInfo) ([]*SpringBootApp, error)
}

type CredentialProvider interface {
	GetCredentials() ([]*Credential, error)
}

type ServerDiscovery interface {
	Prepare() (*Credential, error)
	Server() ServerConnector
	ProcessScan() ([]JavaProcess, error)
	GetTotalMemory() (int64, error)
	GetOsName() (string, error)
	GetOsVersion() (string, error)
	ReadJarFile(location string, walkers ...JarFileWalker) (JarFile, error)
	Finish() error
}

type JarFile interface {
	GetLocation() string
	GetAppType() AppType
	GetArtifactGroup() (string, error)
	GetArtifactName() (string, error)
	GetArtifactVersion() (string, error)
	GetAppName(process JavaProcess) (string, error)
	GetAppPort(process JavaProcess) (int, error)
	GetChecksum() (string, error)
	GetBuildJdkVersion() (string, error)
	GetSpringBootVersion() (string, error)
	GetDependencies() ([]string, error)
	GetApplicationConfigurations() (map[string]string, error)
	GetLoggingFiles() (map[string]string, error)
	GetCertificates() ([]string, error)
	GetStaticFiles() ([]string, error)
	GetLastModifiedTime() (time.Time, error)
	GetSize() (int64, error)
	GetManifests() map[string]string
	GetMavenProject() *mvnparser.MavenProject
}

type JavaProcess interface {
	GetProcessId() int
	GetUid() int
	GetRuntimeJdkVersion() (string, error)
	LocateJarFile() (string, error)
	GetJavaCmd() (string, error)
	GetJvmOptions() ([]string, error)
	GetEnvironments() ([]string, error)
	GetJvmMemory() (int64, error)
	GetPorts() ([]int, error)
	Executor() ServerDiscovery
}

type ServerConnector interface {
	FQDN() string
	Connect(username, password string) error
	Close() error
	Read(remoteLocation string) (io.ReaderAt, os.FileInfo, error)
	RunCmd(cmd string) (string, error)
	Username() string
}

type ServerConnectorFactory interface {
	Create(ctx context.Context, host string, port int) ServerConnector
}
