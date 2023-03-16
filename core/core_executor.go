package core

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
	"microsoft.com/azure-spring-discovery/api/logging"
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ConsoleOutputFilesMapKey = "CONSOLE_OUTPUT_LOGGING_FILES"
	LoggingFilesMapKey       = "LOGGING_FILES"
	ConsoleOutputPatternKey  = "config.pattern.logging.console_output.patterns"
	ConsoleOutputYamlPathKey = "config.pattern.logging.console_output.yamlpath"
)

var SpringBootAppTypes = AppTypes{
	SpringBootFatJar,
	SpringBootThinJar,
	SpringBootExploded,
}

var (
	CrNamePattern          = regexp.MustCompile("[^a-zA-Z0-9\\-\\.]+")
	MavenPomVersionPattern = regexp.MustCompile("-?[0-9\\.]+.*\\.jar")
)

type DiscoveryCoreExecutor struct {
	credentialProvider  CredentialProvider
	targetServerFactory TargetServerFactory
}

func NewCoreExecutor(credentialProvider CredentialProvider,
	targetServerFactory TargetServerFactory,
) *DiscoveryCoreExecutor {
	return &DiscoveryCoreExecutor{
		credentialProvider:  credentialProvider,
		targetServerFactory: targetServerFactory,
	}
}

func (s *DiscoveryCoreExecutor) Discover(ctx context.Context, server *v1alpha1.SpringBootServer) ([]*v1alpha1.SpringBootApp, error) {
	azureLogger := logging.GetAzureLogger(ctx)
	var err error
	if server.Spec.Port == 0 {
		return nil, DiscoveryError{message: fmt.Sprintf("port should not be zero"), severity: v1alpha1.Error}
	}

	discovery := NewLinuxDiscovery(
		ctx,
		s.targetServerFactory.Create(ctx, server.Spec.Server, server.Spec.Port),
		s.credentialProvider,
	)
	var cred *Credential
	if cred, err = discovery.Prepare(); err != nil {
		azureLogger.Error(err, "error connect to server")
		return nil, err
	}
	azureLogger.Info("connect to server successfully", "credential", cred.FriendlyName, "runAsAccountId", cred.Id)
	server.Status.RunAsAccountId = cred.Id
	defer discovery.Finish()

	var processes []JavaProcess
	processes, err = discovery.ProcessScan()
	if err != nil {
		return nil, err
	}
	azureLogger.Info("process scanned", "length", len(processes), "server", server.Name)

	var instanceMap = make(map[string]*v1alpha1.SpringBootApp)
	var errs []error
	for _, process := range processes {
		azureLogger.Info("begin to discover process", "processId", process.GetProcessId(), "server", server.Name)
		var app *v1alpha1.SpringBootApp
		var errInLoop error

		var jarLocation string
		jarLocation, errInLoop = process.LocateJarFile()
		if errInLoop != nil {
			azureLogger.Debug("locate jar file failed", "err", errInLoop)
			errs = append(errs, errInLoop)
			continue
		}

		if existsApp, exists := instanceMap[jarLocation]; exists {
			azureLogger.Debug("jar file already discovered, skip discovery...")
			existsApp.Reference(server)
			continue
		}

		var jar JarFile
		jar, errInLoop = process.Executor().ReadJarFile(jarLocation, DefaultJarFileWalkers...)
		if err != nil {
			azureLogger.Debug("read jar file failed", "err", errInLoop, "location", jarLocation)
			errs = append(errs, errInLoop)
			continue
		}

		app, errInLoop = s.discoverApp(process, jar)
		if errInLoop != nil {
			azureLogger.Debug("discover app failed", "err", errInLoop, "location", jarLocation, "process", process.GetProcessId())
			errs = append(errs, errInLoop)
			continue
		}

		app.Namespace = server.Namespace
		app.Spec.SiteName = server.Spec.SiteName
		app.Spec.LastUpdatedTime = metav1.Now()
		app.Reference(server)
		instanceMap[jarLocation] = app
		azureLogger.Info("finished to discover process, found app", "processId", process.GetProcessId(), "app", app.Name)
	}
	return mapValues(instanceMap), Join(errs...)
}

var getAppName StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppName(process)).OrElse(app.Spec.ArtifactName).OrElse(appDefaultName(process)).Spec("AppName")
}

var getArtifactName StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetArtifactName()).Map(wrap(sanitizeArtifactName)).Spec("ArtifactName")
}

var getAppPort StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppPort(process)).Spec("AppPort")
}

var getChecksum StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetChecksum()).Spec("Checksum")
}

var getJarLocation StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLocation()).Spec("JarFileLocation")
}

var getBuildJdkVersion StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetBuildJdkVersion()).Spec("BuildJdkVersion")
}

var getSpringBootVersion StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetSpringBootVersion()).Spec("SpringBootVersion")
}

var getDependencies StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetDependencies()).Spec("Dependencies")
}

var getCertificates StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetCertificates()).Spec("Certificates")
}

var getAppType StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetAppType()).Map(wrap(mapToAppTypeString)).Spec("AppType")
}

var getStaticContentLocation StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetStaticFiles()).Map(wrap(mapToCommonParentFolder)).Spec("StaticContentLocations")
}

var getApplicationConfigurations StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetApplicationConfigurations()).Map(wrap(mapToKv)).Map(wrap(maskKvSlice)).Spec("ApplicationConfigurations")
}

var getLoggingConfigurations StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLoggingFiles()).Map(wrap(mapToConsoleOutputKV)).Spec("Miscs")
}

var getRuntimeJdkVersion StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetRuntimeJdkVersion()).Spec("RuntimeJdkVersion")
}

var getJvmMemoryInMB StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetJvmMemoryInMb()).Map(wrap(func(f float64) int { return int(f) })).Spec("JvmMemoryInMB")
}

var getEnvironments StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetEnvironments()).Map(wrap(maskSlice)).Spec("Environments")
}

var getJvmOptions StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetJvmOptions()).Map(wrap(maskSlice)).Spec("JvmOptions")
}

var getBindingPorts StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(process.GetPorts()).Spec("BindingPorts")
}

var getLastModifiedTime StepFunc = func(app *v1alpha1.SpringBootApp, process JavaProcess, jarFile JarFile) *Monad {
	return Of(jarFile.GetLastModifiedTime()).Map(wrap(mapToK8sTime)).Spec("LastModifiedTime")
}

func (s *DiscoveryCoreExecutor) discoverApp(process JavaProcess, jar JarFile) (*v1alpha1.SpringBootApp, error) {
	app := &v1alpha1.SpringBootApp{}
	m := NewMonadic(process, jar, app)

	m.
		Then(getArtifactName). // get artifact name should be first otherwise not able to use artifact name as fallback of app name
		Then(getAppName).
		Then(getChecksum).
		Then(getAppPort).
		Then(getJarLocation).
		Then(getBuildJdkVersion).
		Then(getRuntimeJdkVersion).
		Then(getSpringBootVersion).
		Then(getDependencies).
		Then(getCertificates).
		Then(getAppType).
		Then(getStaticContentLocation).
		Then(getApplicationConfigurations).
		Then(getLoggingConfigurations).
		Then(getJvmMemoryInMB).
		Then(getEnvironments).
		Then(getJvmOptions).
		Then(getBindingPorts).
		Then(getLastModifiedTime)

	if m.err != nil {
		return nil, m.err
	}

	app.Name = app.Spec.Checksum

	return app, nil
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
		if !slices.Contains(folders, "/"+base) {
			folders = append(folders, "/"+base)
		}
	}
	return folders
}

func maskSlice(origin []string) []string {
	if len(origin) == 0 {
		return nil
	}
	propsMasker := NewPropsMasker()
	content := strings.Join(origin, "\n")
	masked, err := propsMasker.Mask(content)
	if err != nil {
		logf.Log.Error(err, "unable to mask slices")
		return nil
	}

	return strings.Split(masked, "\n")
}

func maskKvSlice(kvs []*v1alpha1.KV) []*v1alpha1.KV {
	if len(kvs) == 0 {
		return nil
	}
	yamlMasker := NewYamlMasker()
	propsMasker := NewPropsMasker()

	var masked string
	var err error
	for _, kv := range kvs {
		ext := filepath.Ext(kv.Key)
		switch ext {
		case ".yml", ".yaml":
			masked, err = yamlMasker.Mask(kv.Value)
		case ".properties":
			masked, err = propsMasker.Mask(kv.Value)
		default:
			masked = kv.Value
		}
		if err != nil {
			logf.Log.Error(err, "failed to mask configuration files", "file", kv.Key)
			return nil
		}

		kv.Value = masked
	}

	return kvs
}

func mapToKv(m map[string]string) []*v1alpha1.KV {
	if m == nil {
		return nil
	}
	var kv []*v1alpha1.KV
	for k, v := range m {
		kv = append(kv, &v1alpha1.KV{Key: k, Value: v})
	}
	return kv
}

func mapToConsoleOutputKV(loggingFiles map[string]string) []*v1alpha1.KV {
	return []*v1alpha1.KV{
		{Key: ConsoleOutputFilesMapKey, Value: strings.Join(filterMap(loggingFiles, hasConsoleOutput), ",")},
		{Key: LoggingFilesMapKey, Value: strings.Join(mapKeys(loggingFiles), ",")},
	}
}

func mapToK8sTime(time time.Time) metav1.Time {
	return metav1.NewTime(time)
}

func hasConsoleOutput(filename, content string) bool {
	if slices.Contains([]string{".json", ".jsn", ".yaml", ".yml"}, filepath.Ext(filename)) {
		var node yaml.Node
		err := yaml.Unmarshal([]byte(content), &node)
		if err != nil {
			klog.Info("encounter error %s to unmarshall logging file %s", err.Error(), filename)
			return false
		}
		for _, path := range Patterns.ConsoleOutputYamlPatterns {
			if findings, err := path.Find(&node); err != nil {
				klog.Info("encounter error %s to find console output in logging file %s", err.Error(), filename)
				continue
			} else {
				if len(findings) > 0 {
					return true
				}
			}
		}
	} else {
		for _, p := range Patterns.ConsoleOutputRegexPatterns {
			if p.MatchString(content) {
				return true
			}
		}
	}
	return false
}

func sanitizeArtifactName(artifactName string) string {
	return MavenPomVersionPattern.ReplaceAllString(artifactName, "")
}

func markStatus(server *v1alpha1.SpringBootServer, err error) {
	server.Status.ObservedGeneration = server.Generation
	if err != nil {
		server.Status.ProvisioningStatus = v1alpha1.ProvisioningStatus{
			Status: v1alpha1.Failed,
			Error:  mapError(err, server.Status.RunAsAccountId).ProvisioningError(),
		}
		server.Status.Errors = mapErrors(server.Status.RunAsAccountId, err)
	} else {
		server.Status.ProvisioningStatus = v1alpha1.ProvisioningStatus{
			Status: v1alpha1.Succeeded,
		}
	}
	server.Status.ProvisioningStatus.OperationID = server.Annotations[v1alpha1.AnnotationsOperationId]
}

func appDefaultName(process JavaProcess) string {
	return fmt.Sprintf("app-%d", process.GetProcessId())
}

func MergeApp(a, b *v1alpha1.SpringBootApp) {
	a.Merge(b)
}
