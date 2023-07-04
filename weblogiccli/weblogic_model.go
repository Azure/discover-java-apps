package main

import (
	"github.com/Azure/discover-java-apps/weblogic"
	"time"
)

type WeblogicCliApp struct {
	Server            string `json:"server" csv:"Server"`
	AppName           string `json:"appName" csv:"AppName"`
	AppType           string `json:"appType" csv:"AppType"`
	AppPort           int    `json:"appPort" csv:"AppPort"`
	JvmMemory         int64  `json:"jvmMemoryInMB" csv:"JvmHeapMemory(MB)"`
	OsName            string `json:"osName" csv:"OsName"`
	OsVersion         string `json:"osVersion" csv:"OsVersion"`
	JarFileLocation   string `json:"jarFileLocation" csv:"JarFileLocation"`
	LastModifiedTime  string `json:"lastModifiedTime" csv:"JarFileModifiedTime"`
	WeblogicVersion   string `json:"weblogicVersion" csv:"WeblogicVersion"`
	WeblogicPatches   string `json:"weblogicPatches" csv:"WeblogicPatches"`
	DeploymentTarget  string `json:"deploymentTarget" csv:"DeploymentTarget"`
	RuntimeJdkVersion string `json:"runtimeJdkVersion" csv:"RuntimeJdkVersion"`
	ServerType        string `json:"serverType" csv:"ServerType"`
	OracleHome        string `json:"oracleHome" csv:"OracleHome"`
	DomainHome        string `json:"domainHome" csv:"DomainHome"`
}

type weblogicAppConverter struct {
}

type Converter[From any, To any] interface {
	Convert(from From) To
}

func NewWeblogicAppConverter() Converter[[]*weblogic.WeblogicApp, []*WeblogicCliApp] {
	return &weblogicAppConverter{}
}

func (s weblogicAppConverter) Convert(apps []*weblogic.WeblogicApp) []*WeblogicCliApp {
	var results []*WeblogicCliApp

	for _, app := range apps {

		var appType = "war"

		results = append(results, &WeblogicCliApp{
			Server:            app.Server,
			OsName:            app.OsName,
			OsVersion:         app.OsVersion,
			JvmMemory:         app.JvmMemoryInMB / weblogic.MiB,
			DeploymentTarget:  app.DeploymentTarget,
			AppName:           app.AppName,
			AppType:           appType,
			ServerType:        app.ServerType,
			AppPort:           app.AppPort,
			WeblogicVersion:   app.WeblogicVersion,
			WeblogicPatches:   app.WeblogicPatches,
			JarFileLocation:   app.AppFileLocation,
			RuntimeJdkVersion: app.RuntimeJdkVersion,
			OracleHome:        app.OracleHome,
			DomainHome:        app.DomainHome,

			LastModifiedTime: app.LastModifiedTime.UTC().Format(time.RFC3339),
		})
	}
	return results
}
