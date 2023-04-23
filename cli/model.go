package main

import (
	"github.com/Azure/discover-java-apps/springboot"
	"time"
)

type CliApp struct {
	Server            string `json:"server" csv:"Server"`
	AppName           string `json:"appName" csv:"AppName"`
	AppType           string `json:"appType" csv:"AppType"`
	AppPort           int    `json:"appPort" csv:"AppPort"`
	ArtifactGroup     string `json:"artifactGroup" csv:"MavenArtifactGroup"`
	ArtifactName      string `json:"artifactName" csv:"MavenArtifact"`
	ArtifactVersion   string `json:"artifactVersion" csv:"MavenArtifactVersion"`
	SpringBootVersion string `json:"springBootVersion" csv:"SpringBootVersion"`
	BuildJdkVersion   string `json:"buildJdkVersion" csv:"BuildJdkVersion"`
	RuntimeJdkVersion string `json:"runtimeJdkVersion" csv:"RuntimeJdkVersion"`
	JvmMemory         int64  `json:"jvmMemoryInMB" csv:"JvmHeapMemory(MB)"`
	OsName            string `json:"osName" csv:"OsName"`
	OsVersion         string `json:"osVersion" csv:"OsVersion"`
	JarFileLocation   string `json:"jarFileLocation" csv:"JarFileLocation"`
	JarSize           int64  `json:"jarSizeInKB" csv:"JarFileSize(KB)"`
	LastModifiedTime  string `json:"lastModifiedTime" csv:"JarFileModifiedTime"`
}

type Converter[From any, To any] interface {
	Convert(from From) To
}

type springBootAppConverter struct {
}

func NewSpringBootAppConverter() Converter[[]*springboot.SpringBootApp, []*CliApp] {
	return &springBootAppConverter{}
}

func (s springBootAppConverter) Convert(apps []*springboot.SpringBootApp) []*CliApp {
	var results []*CliApp

	for _, app := range apps {

		var appType = "ExecutableJar"

		if springboot.SpringBootAppTypes.Contains(app.AppType) {
			appType = "SpringBoot"
		}

		results = append(results, &CliApp{
			Server:            app.Runtime.Server,
			AppName:           app.AppName,
			AppType:           appType,
			AppPort:           app.Runtime.AppPort,
			ArtifactGroup:     app.Artifact.Group,
			ArtifactName:      app.Artifact.Name,
			ArtifactVersion:   app.Artifact.Version,
			SpringBootVersion: app.SpringBootVersion,
			BuildJdkVersion:   app.BuildJdkVersion,
			RuntimeJdkVersion: app.Runtime.RuntimeJdkVersion,
			JarFileLocation:   app.JarFileLocation,
			OsName:            app.Runtime.OsName,
			OsVersion:         app.Runtime.OsVersion,
			JarSize:           app.JarSize / springboot.KiB,
			JvmMemory:         app.Runtime.JvmMemory / springboot.MiB,
			LastModifiedTime:  app.LastModifiedTime.UTC().Format(time.RFC3339),
		})
	}
	return results
}
