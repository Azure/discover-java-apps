package weblogic

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"strconv"
	"strings"
)

type webLogicDiscoveryExecutor struct {
	credentialProvider     CredentialProvider
	serverConnectorFactory ServerConnectorFactory
}

func NewWeblogicDiscoveryExecutor(
	credentialProvider CredentialProvider,
	serverConnectorFactory ServerConnectorFactory,
) DiscoveryExecutor {
	return &webLogicDiscoveryExecutor{
		credentialProvider:     credentialProvider,
		serverConnectorFactory: serverConnectorFactory,
	}
}

func (s *webLogicDiscoveryExecutor) Discover(ctx context.Context, serverConnectionInfo ServerConnectionInfo, alternativeConnectionInfos ...ServerConnectionInfo) ([]*WeblogicApp, error) {
	azureLogger := GetAzureLogger(ctx)
	azureLogger.Info("going to discover weblogic apps")
	var err error
	serverDiscovery, cred, err := s.tryConnect(ctx, append([]ServerConnectionInfo{serverConnectionInfo}, alternativeConnectionInfos...))
	if err != nil {
		return nil, err
	}
	azureLogger.Info("connect to serverConnectionInfo successfully", "cred", cred.FriendlyName, "runAsAccountId", cred.Id)
	defer serverDiscovery.Finish()

	var processes []WebLogicProcess
	processes, err = serverDiscovery.ProcessScan()
	if err != nil {
		return nil, err
	}
	azureLogger.Info("weblogic process scanned", "length", len(processes))

	var apps []*WeblogicApp
	var errs []error
	if len(processes) > 0 {
		process := processes[0]
		azureLogger.Info("begin to discover process", "processId", process.GetProcessId())

		tempfolderPath := process.CreateTempFolder()
		println("tempfolderPath: " + tempfolderPath)
		defer process.DeleteTempFolder(tempfolderPath)

		weblogicName, _ := process.GetWeblogicName()
		println("WeblogicServerName: " + weblogicName)

		javaHome, _ := process.GetJavaHome()
		println("Java_Home: " + javaHome)

		// Get weblogic Home
		weblogicHome, _ := process.GetWeblogicHome()
		println("Weblogic_Home: " + weblogicHome)
		oracleHome := getOracleHome(weblogicHome)
		println("Oracle_Home: " + oracleHome)

		connection := buildConnection(cred.WeblogicUsername, cred.WeblogicPassword, cred.Weblogicport)
		domainHome, _ := process.GetDomainHome(oracleHome, tempfolderPath, connection)
		println("The Domain_Home is: " + domainHome)
		applicationPathMap := process.GetApplicationsAndPath(domainHome, tempfolderPath, connection)
		print("The total number of applications detected is: ")
		println(len(applicationPathMap))

		if len(applicationPathMap) > 0 {
			process.UploadAndInstallWDT(tempfolderPath)
			process.RunDiscoverDomainCommand(tempfolderPath, javaHome, cred.WeblogicUsername, cred.WeblogicPassword, oracleHome, cred.Weblogicport)
			discoverDomainResult := process.GetDiscoverDomainResult(tempfolderPath)
			// Iterate over the map
			for application, app_path := range applicationPathMap {
				fmt.Printf("application: %s, app_path: %s\n", application, app_path)
				var app = new(WeblogicApp)

				app.Server = getServer(process)

				app.OsName, _ = process.Executor().GetOsName()
				app.OsVersion, _ = process.Executor().GetOsVersion()
				memory, _ := getJvmMemory(process)
				app.JvmMemoryInMB = memory
				app.RuntimeJdkVersion, _ = process.GetRuntimeJdkVersion()
				app.AppFileLocation = app_path
				app.LastModifiedTime, _ = process.GetLastModifiedTime(app_path)

				app.AppName = application
				app.AppType = getAppType(discoverDomainResult, application)
				app.DeploymentTarget = getDeploymentTarget(discoverDomainResult, application)
				app.ServerType = getServerType(discoverDomainResult)
				app.AppPort = getAppPort(discoverDomainResult)

				app.WeblogicVersion = process.GetWeblogicVersion(domainHome)
				app.WeblogicPatches = process.GetWeblogicPatch(domainHome)

				app.OracleHome = oracleHome
				app.DomainHome = domainHome

				azureLogger.Info("finished to discover process, found app", "processId", process.GetProcessId(), "app", app.AppName)

				apps = append(apps, app)
			}
		}

	} else {
		println("No weblogic process detected")
	}

	println("------------------------------------------------------")
	return apps, Join(errs...)
}

func getOracleHome(home string) string {
	return strings.TrimSuffix(home, "/wlserver/server")
}

func buildConnection(weblogicuserName, weblogicPassword string, port int) string {
	return "connect('" + weblogicuserName + "','" + weblogicPassword + "','t3://localhost:" + strconv.Itoa(port) + "')"
}

func (s *webLogicDiscoveryExecutor) tryConnect(ctx context.Context, serverConnectionInfos []ServerConnectionInfo) (ServerDiscovery, *Credential, error) {
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
		)

		if cred, err = serverDiscovery.Prepare(); err != nil {
			_ = serverDiscovery.Finish()
			if IsCredentialError(err) {
				// if credential error, the server is connectable, just break to avoid nonsense try
				break
			}
			azureLogger.Warning(err, "failed to connect to", "server", info.Server)
			continue
		} else {
			return serverDiscovery, cred, nil
		}
	}

	// every connection info has been tried, but still failed
	return nil, cred, err
}

var getAppPort = func(discoverDomainResult string) int {
	value, _ := getYamlValue(discoverDomainResult, "topology.Server.admin.WebServer.FrontendHTTPPort")
	return value.(int)
}

var getAppType = func(discoverDomainResult string, application string) string {
	value, _ := getYamlValue(discoverDomainResult, "appDeployments.Application."+application+".ModuleType")
	return value.(string)
}

var getDeploymentTarget = func(discoverDomainResult string, application string) string {
	value, _ := getYamlValue(discoverDomainResult, "appDeployments.Application."+application+".Target")
	return value.(string)
}

var getServerType = func(discoverDomainResult string) string {
	return "weblogic"
}

var getArtifactName = func(process WebLogicProcess, discoverDomainResult string, application string) string {
	value, _ := getYamlValue(discoverDomainResult, "appDeployments.Application."+application+".SourcePath")
	return value.(string)
}

func getYamlValue(yamlString string, path string) (interface{}, bool) {
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlString), &data); err != nil {
		log.Fatal(err)
	}
	return getNestedValue(data, path)
}

func getNestedValue(data map[string]interface{}, path string) (interface{}, bool) {
	keys := splitPath(path)

	for _, key := range keys {
		value, ok := data[key]
		if !ok {
			return nil, false
		}

		if nestedData, ok := value.(map[string]interface{}); ok {
			data = nestedData
		} else {
			return value, true
		}
	}

	return nil, false
}

func splitPath(path string) []string {
	return strings.Split(path, ".")
}

var getServer = func(process WebLogicProcess) string {
	return process.Executor().Server().FQDN()
}

var getJvmMemory = func(process WebLogicProcess) (int64, error) {
	return process.GetJvmMemory()
}
