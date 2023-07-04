package weblogic

import (
	"context"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"time"
)

type AppType string

type AppTypes []AppType

type ServerConnectionInfo struct {
	Server string
	Port   int
}

type Credential struct {
	Id               string `json:"Id,omitempty"`
	FriendlyName     string `json:"FriendlyName,omitempty"`
	Username         string `json:"UserName,omitempty"`
	Password         string `json:"Password,omitempty"`
	CredentialType   string `json:"CredentialType,omitempty"`
	WeblogicUsername string `json:"WeblogicUsername,omitempty"`
	WeblogicPassword string `json:"WeblogicPassword,omitempty"`
	Weblogicport     int    `json:"Weblogicport,omitempty"`
}

type CredentialProvider interface {
	GetCredentials() ([]*Credential, error)
}

type WebLogicProcess interface {
	GetProcessId() int
	GetUid() int
	GetRuntimeJdkVersion() (string, error)
	GetJavaCmd() (string, error)
	GetJvmOptions() ([]string, error)
	GetJvmMemory() (int64, error)
	GetPorts() ([]int, error)
	UploadAndInstallWDT(Randomfolder string) string
	GetApplicationsAndPath(DomainHome string, Randomfolder string, connection string) map[string]string
	GetWeblogicVersion(DomainHome string) string
	GetWeblogicPatch(DomainHome string) string
	GetLastModifiedTime(path string) (time.Time, error)
	RunDiscoverDomainCommand(randomfolder, javaHome, weblogicUser, weblogicPassword, oracleHome string, port int) string
	GetDomainHome(Oracle_Home string, Randomfolder string, connection string) (string, error)
	GetWeblogicName() (string, error)
	GetWeblogicHome() (string, error)
	GetJavaHome() (string, error)
	GetDiscoverDomainResult(Randomfolder string) string
	CreateTempFolder() string
	DeleteTempFolder(path string) string
	Executor() ServerDiscovery
}

type ServerDiscovery interface {
	Prepare() (*Credential, error)
	Server() ServerConnector
	ProcessScan() ([]WebLogicProcess, error)
	GetTotalMemory() (int64, error)
	GetOsName() (string, error)
	GetOsVersion() (string, error)
	Finish() error
}

type ServerConnector interface {
	FQDN() string
	Connect(username, password string) error
	Close() error
	Read(remoteLocation string) (io.ReaderAt, os.FileInfo, error)
	RunCmd(cmd string) (string, error)
	Username() string
	Client() *ssh.Client
}

type Artifact struct {
	Group   string `json:"group"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerConnectorFactory interface {
	Create(ctx context.Context, host string, port int) ServerConnector
}

type WeblogicApp struct {
	Server             string    `json:"server"`
	AppName            string    `json:"appName"`
	AppType            string    `json:"appType"`
	Artifact           *Artifact `json:"artifact"`
	ArtifactName       string    `json:"artifactName"`
	ServerType         string    `json:"servertype"`
	DeploymentTarget   string    `json:"deploymentTarget"`
	WeblogicVersion    string    `json:"weblogicVersion"`
	WeblogicPatches    string    `json:"weblogicPatches"`
	Comment1           string    `json:"_comment1"`
	WeblogicServerName string    `json:"weblogicServerName"`
	RuntimeJdkVersion  string    `json:"runtimeJdkVersion"`
	OsName             string    `json:"osName"`
	OsVersion          string    `json:"osVersion"`
	AppFileLocation    string    `json:"appFileLocation"`
	JvmMemoryInMB      int64     `json:"jvmMemoryInMB"`
	AppPort            int       `json:"appPort"`
	ContextRoot        string    `json:"contextRoot"`
	LastModifiedTime   time.Time `json:"lastModifiedTime"`
	OracleHome         string    `json:"oracleHome"`
	DomainHome         string    `json:"domainHome"`
}

type DiscoveryExecutor interface {
	Discover(ctx context.Context, server ServerConnectionInfo, alternativeConnectionInfos ...ServerConnectionInfo) ([]*WeblogicApp, error)
}
