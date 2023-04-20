package springboot

import (
	"context"
	errors "github.com/pkg/errors"
	"reflect"
	"strings"
	"time"
)

type springBootDiscoveryExecutor struct {
	credentialProvider     CredentialProvider
	serverConnectorFactory ServerConnectorFactory
	cfg                    YamlConfig
}

func NewSpringBootDiscoveryExecutor(
	credentialProvider CredentialProvider,
	serverConnectorFactory ServerConnectorFactory,
	cfg YamlConfig,
) DiscoveryExecutor {
	return &springBootDiscoveryExecutor{
		credentialProvider:     credentialProvider,
		serverConnectorFactory: serverConnectorFactory,
		cfg:                    cfg,
	}
}

func (s *springBootDiscoveryExecutor) Discover(ctx context.Context, serverConnectionInfo ServerConnectionInfo, alternativeConnectionInfos ...ServerConnectionInfo) ([]*SpringBootApp, error) {
	azureLogger := GetAzureLogger(ctx)
	azureLogger.Info("going to discover")
	var err error
	serverDiscovery, cred, err := s.tryConnect(ctx, append(alternativeConnectionInfos, serverConnectionInfo))
	if err != nil {
		return nil, errors.Wrap(err, "connection failed")
	}
	azureLogger.Info("connect to serverConnectionInfo successfully", "credential", cred.FriendlyName, "runAsAccountId", cred.Id)
	defer serverDiscovery.Finish()

	var processes []JavaProcess
	processes, err = serverDiscovery.ProcessScan()
	if err != nil {
		return nil, err
	}
	azureLogger.Info("process scanned", "length", len(processes))

	var jarCache = make(map[string]JarFile)
	var apps []*SpringBootApp
	var errs []error
	for _, process := range processes {
		azureLogger.Info("begin to discover process", "processId", process.GetProcessId())
		var app *SpringBootApp
		var errInLoop error

		var jarLocation string
		jarLocation, errInLoop = process.LocateJarFile()
		if errInLoop != nil {
			azureLogger.Warning(errInLoop, "locate jar file failed")
			errs = append(errs, errInLoop)
			continue
		}

		var jar JarFile
		var exists bool
		if jar, exists = jarCache[jarLocation]; exists {
			azureLogger.Debug("jar file already discovered")
		} else {
			jar, errInLoop = process.Executor().ReadJarFile(jarLocation, DefaultJarFileWalkers...)
			if errInLoop != nil {
				azureLogger.Error(errInLoop, "read jar file failed", "location", jarLocation, "type", reflect.TypeOf(errInLoop))
				errs = append(errs, errInLoop)
				continue
			}
		}

		app, errInLoop = s.discoverApp(process, jar)
		if errInLoop != nil {
			azureLogger.Warning(errInLoop, "discover app failed", "location", jarLocation, "process", process.GetProcessId(), "type", reflect.TypeOf(errInLoop))
			errs = append(errs, errInLoop)
			continue
		}
		jarSize, _ := jar.GetSize()
		app.LastUpdatedTime = time.Now()
		app.JarSize = jarSize
		jarCache[jarLocation] = jar

		azureLogger.Info("finished to discover process, found app", "processId", process.GetProcessId(), "app", app.AppName)

		apps = append(apps, app)
	}

	return apps, Join(errs...)
}

var getAppName StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppName(process)).Field("AppName")
}

var getArtifactName StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetArtifactName()).Map(wrap(sanitizeArtifactName)).Field("Name")
}

var getArtifactGroup StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetArtifactGroup()).Field("Group")
}

var getArtifactVersion StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetArtifactVersion()).Field("Version")
}

var getAppPort StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppPort(process)).Field("AppPort")
}

var getJavaCmd StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetJavaCmd()).Field("JavaCmd")
}

var getServer StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.Executor().Server().FQDN()).Field("Server")
}

var getChecksum StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetChecksum()).Field("Checksum")
}

var getJarLocation StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLocation()).Field("JarFileLocation")
}

var getBuildJdkVersion StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetBuildJdkVersion()).Field("BuildJdkVersion")
}

var getSpringBootVersion StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetSpringBootVersion()).Field("SpringBootVersion")
}

var getDependencies StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetDependencies()).Field("Dependencies")
}

var getCertificates StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetCertificates()).Field("Certificates")
}

var getAppType StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppType()).Field("AppType")
}

var getStaticContentLocation StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetStaticFiles()).Map(wrap(mapToCommonParentFolder)).Field("StaticContentLocations")
}

var getApplicationConfigurations StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetApplicationConfigurations()).Field("ApplicationConfigurations")
}

var getLoggingConfigurations StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLoggingFiles()).Field("LoggingConfigurations")
}

var getRuntimeJdkVersion StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetRuntimeJdkVersion()).Field("RuntimeJdkVersion")
}

var getJvmMemory StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetJvmMemory()).Field("JvmMemory")
}

var getEnvironments StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetEnvironments()).Field("Environments")
}

var getJvmOptions StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetJvmOptions()).Field("JvmOptions")
}

var getBindingPorts StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetPorts()).Field("BindingPorts")
}

var getLastModifiedTime StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLastModifiedTime()).Field("LastModifiedTime")
}

var getOsName StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.Executor().GetOsName()).Field("OsName")
}

var getOsVersion StepFunc = func(process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.Executor().GetOsVersion()).Field("OsVersion")
}

func (s *springBootDiscoveryExecutor) discoverApp(process JavaProcess, jar JarFile) (*SpringBootApp, error) {

	m := NewMonadic[*SpringBootApp](process, jar)
	app, err := m.
		Apply(getAppName).
		Apply(getChecksum).
		Apply(getJarLocation).
		Apply(getBuildJdkVersion).
		Apply(getSpringBootVersion).
		Apply(getDependencies).
		Apply(getCertificates).
		Apply(getAppType).
		Apply(getStaticContentLocation).
		Apply(getApplicationConfigurations).
		Apply(getLoggingConfigurations).
		Apply(getLastModifiedTime).
		Get()

	if err != nil {
		return nil, err
	}

	am := NewMonadic[*Artifact](process, jar)
	artifact, err := am.
		Apply(getArtifactName).
		Apply(getArtifactGroup).
		Apply(getArtifactVersion).
		Get()

	if err != nil {
		return nil, err
	}
	app.Artifact = artifact

	runtime, err := NewMonadic[*Runtime](process, jar).
		Apply(getAppPort).
		Apply(getJavaCmd).
		Apply(getServer).
		Apply(getRuntimeJdkVersion).
		Apply(getJvmMemory).
		Apply(getEnvironments).
		Apply(getJvmOptions).
		Apply(getBindingPorts).
		Apply(getOsName).
		Apply(getOsVersion).
		Get()
	if err != nil {
		return nil, err
	}

	app.Runtime = runtime

	if err != nil {
		return nil, err
	}

	return app, err
}

func mapToAppTypeString(appType AppType) string {
	return string(appType)
}

func mapToCommonParentFolder(origin []string) []string {
	if origin == nil {
		return nil
	}

	var folders []string
	for _, f := range origin {
		splits := strings.Split(f, "/")
		base := splits[0]
		if !Contains(folders, "/"+base) {
			folders = append(folders, "/"+base)
		}
	}
	return folders
}

func sanitizeArtifactName(artifactName string) string {
	return Patterns.MavenPomVersionPattern.ReplaceAllString(artifactName, "")
}

func (s *springBootDiscoveryExecutor) tryConnect(ctx context.Context, serverConnectionInfos []ServerConnectionInfo) (ServerDiscovery, *Credential, error) {
	azureLogger := GetAzureLogger(ctx)
	var serverDiscovery ServerDiscovery
	var cred *Credential
	var err error
	for _, info := range serverConnectionInfos {
		if len(info.Server) == 0 || info.Port == 0 {
			azureLogger.Warning(err, "invalid connection info", "server", info.Server, "port", info.Port)
			continue
		}
		serverDiscovery = NewLinuxServerDiscovery(
			ctx,
			s.serverConnectorFactory.Create(ctx, info.Server, info.Port),
			s.credentialProvider,
			s.cfg,
		)

		if cred, err = serverDiscovery.Prepare(); err != nil {
			azureLogger.Warning(err, "failed to connect to", "server", info.Server)
			continue
		} else {
			return serverDiscovery, cred, nil
		}
	}

	// every connection info has been tried, but still failed
	return nil, nil, err

}
