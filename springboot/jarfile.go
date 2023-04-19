package springboot

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/xml"
	"github.com/creekorful/mvnparser"
	"github.com/pkg/errors"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	AppNameField                = "Implementation-Title"
	VersionField                = "Implementation-Version"
	MainClassField              = "Main-Class"
	JdkVersionField             = "Build-Jdk-Spec"
	JdkVersionFieldFor1x        = "Build-Jdk"
	SpringBootVersionField      = "Spring-Boot-Version"
	JarLauncherClassName        = "org.springframework.boot.loader.JarLauncher"
	PropertiesLauncherClassName = "org.springframework.boot.loader.PropertiesLauncher"

	PomFileName                 = "pom.xml"
	SpringBootStarterGroupId    = "org.springframework.boot"
	SpringBootStarterArtifactId = "spring-boot-starter-parent"
	SpringBootJarFilePrefix     = "spring-boot"
	JavaVersionPropertyName     = "java.version"
	CompilerTargetPropertyName  = "maven.compiler.target"
	CompilerReleasePropertyName = "maven.compiler.target"
	DefaultClasspath            = "BOOT-INF/classes/"
	DefaultLibPath              = "BOOT-INF/lib/"
	DefaultMvnPath              = "META-INF/maven/"
	ApplicationNameKey          = "spring.application.name"
	ApplicationPortKey          = "server.port"
)

type jarFile struct {
	checksum                  string
	remoteLocation            string
	manifests                 map[string]string
	dependencies              []string
	applicationConfigurations map[string]string
	loggingConfigs            map[string]string
	certificates              []string
	staticFiles               []string
	mvnProject                *mvnparser.MavenProject
	lastModifiedTime          time.Time
	size                      int64
}

func (j *jarFile) GetAppType() AppType {
	var zero AppType
	var tryManifest tryFunc[*jarFile, AppType] = func(j *jarFile) (AppType, bool) {
		if value, ok := j.manifests[MainClassField]; ok {
			switch value {
			case JarLauncherClassName:
			case PropertiesLauncherClassName:
				return SpringBootFatJar, true
			default:
				return ExecutableJar, true
			}
		}

		return zero, false
	}

	var tryPom tryFunc[*jarFile, AppType] = func(file *jarFile) (AppType, bool) {
		if j.mvnProject != nil && j.mvnProject.Parent.GroupId == SpringBootStarterGroupId && j.mvnProject.Parent.ArtifactId == SpringBootStarterArtifactId {
			return SpringBootFatJar, true
		}
		return zero, false
	}

	var tryDeps tryFunc[*jarFile, AppType] = func(file *jarFile) (AppType, bool) {
		for _, lib := range j.dependencies {
			if strings.Contains(lib, SpringBootJarFilePrefix) {
				return SpringBootFatJar, true
			}
		}
		return zero, false
	}

	var funcs = tryFuncs[*jarFile, AppType]{tryManifest, tryPom, tryDeps}
	if value, ok := funcs.try(j); ok {
		return value
	}
	return ExecutableJar
}

func (j *jarFile) GetArtifactGroup() (string, error) {
	return j.mvnProject.GroupId, nil
}

func (j *jarFile) GetArtifactName() (string, error) {

	var tryManifest tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if value, ok := j.manifests[AppNameField]; ok {
			return strings.TrimSpace(value), len(strings.TrimSpace(value)) > 0
		}
		return "", false
	}

	var tryPom tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if j.mvnProject == nil {
			return "", false
		}
		return j.mvnProject.Name, len(j.mvnProject.Name) > 0
	}

	var tryFilename tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		filename := filepath.Base(j.remoteLocation)
		return sanitizeArtifactName(strings.ReplaceAll(filename, filepath.Ext(filename), "")), true
	}

	var funcs = tryFuncs[*jarFile, string]{tryPom, tryManifest, tryFilename}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return "", nil

}

func (j *jarFile) GetArtifactVersion() (string, error) {
	var tryManifest tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if value, ok := j.manifests[VersionField]; ok {
			return strings.TrimSpace(value), true
		}
		return "", false
	}

	var tryPom tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if j.mvnProject == nil {
			return "", false
		}

		if len(j.mvnProject.Version) > 0 {
			return j.mvnProject.Version, true
		}
		return "", false
	}

	var funcs = tryFuncs[*jarFile, string]{tryPom, tryManifest}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return "", nil
}

func (j *jarFile) GetAppName(process JavaProcess) (string, error) {
	var tryAppYaml tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		for f, text := range j.applicationConfigurations {
			ext := filepath.Ext(f)
			if ext == ".yml" || ext == ".yaml" {
				if find, ok := GetConfigFromYaml[string](ApplicationNameKey, text); ok {
					return strings.TrimSpace(find), len(strings.TrimSpace(find)) > 0
				}
			}
		}

		return "", false
	}

	var tryAppProps tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		for f, text := range j.applicationConfigurations {
			ext := filepath.Ext(f)
			if ext == ".properties" {
				if find, ok := GetConfigFromProperties(ApplicationNameKey, text); ok {
					return strings.TrimSpace(find), len(strings.TrimSpace(find)) > 0
				}
			}
		}

		return "", false
	}

	var tryOptions tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		options, err := process.GetJvmOptions()
		if err != nil {
			return "", false
		}
		m := ParseProperties(strings.Join(options, "\n"))
		for k, v := range m {
			if strings.TrimSpace(k) == "-D"+ApplicationNameKey {
				return v, len(v) > 0
			}
		}

		return "", false
	}

	var tryFilename tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		return sanitizeArtifactName(filepath.Base(j.remoteLocation)), true
	}

	var funcs = tryFuncs[*jarFile, string]{tryOptions, tryAppProps, tryAppYaml, tryFilename}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return "", nil
}

func (j *jarFile) GetAppPort(process JavaProcess) (int, error) {

	var tryAppYaml tryFunc[*jarFile, int] = func(j *jarFile) (int, bool) {
		for f, text := range j.applicationConfigurations {
			ext := filepath.Ext(f)
			if ext == ".yml" || ext == ".yaml" {
				if port, ok := GetConfigFromYaml[int](ApplicationPortKey, text); ok {
					return port, true
				} else {
					if find, ok := GetConfigFromYaml[string](ApplicationPortKey, text); ok {
						var port int
						var err error
						if port, err = strconv.Atoi(strings.TrimSpace(find)); err != nil {
							return 0, false
						} else {
							return port, true
						}
					}
				}
			}
		}

		return 0, false
	}

	var tryAppProps tryFunc[*jarFile, int] = func(j *jarFile) (int, bool) {
		for f, text := range j.applicationConfigurations {
			ext := filepath.Ext(f)
			if ext == ".properties" {
				if find, ok := GetConfigFromProperties(ApplicationPortKey, text); ok {
					var port int
					var err error
					if port, err = strconv.Atoi(strings.TrimSpace(find)); err != nil {
						return 0, false
					} else {
						return port, true
					}
				}
			}
		}

		return 0, false
	}

	var tryOptions tryFunc[*jarFile, int] = func(j *jarFile) (int, bool) {
		options, err := process.GetJvmOptions()
		if err != nil {
			return 0, false
		}
		m := ParseProperties(strings.Join(options, "\n"))
		for k, v := range m {
			if k == "-D"+ApplicationPortKey || k == "--"+ApplicationPortKey {
				var port int
				var err error
				if port, err = strconv.Atoi(strings.TrimSpace(v)); err != nil {
					return 0, false
				} else {
					return port, true
				}
			}
		}

		return 0, false
	}

	var defaultPort tryFunc[*jarFile, int] = func(j *jarFile) (int, bool) {
		return 8080, true
	}

	var funcs = tryFuncs[*jarFile, int]{tryOptions, tryAppYaml, tryAppProps, defaultPort}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return 0, nil

}

func (j *jarFile) GetChecksum() (string, error) {
	return j.checksum, nil
}

func (j *jarFile) GetBuildJdkVersion() (string, error) {
	var tryPom tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if j.mvnProject == nil {
			return "", false
		}
		for _, key := range []string{JavaVersionPropertyName, CompilerReleasePropertyName, CompilerTargetPropertyName} {
			if value, ok := j.mvnProject.Properties[key]; ok {
				return strings.TrimSpace(value), len(strings.TrimSpace(value)) > 0
			}
		}

		return "", false
	}

	var try1x tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if value, ok := j.manifests[JdkVersionFieldFor1x]; ok {
			return strings.TrimSpace(value), len(strings.TrimSpace(value)) > 0
		}
		return "", false
	}

	var try2x tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if value, ok := j.manifests[JdkVersionField]; ok {
			return strings.TrimSpace(value), len(strings.TrimSpace(value)) > 0
		}
		return "", false
	}

	var funcs = tryFuncs[*jarFile, string]{tryPom, try2x, try1x}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return "", nil
}

func (j *jarFile) GetSpringBootVersion() (string, error) {
	var tryManifest tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if value, ok := j.manifests[SpringBootVersionField]; ok {
			return strings.TrimSpace(value), true
		}
		return "", false
	}

	var tryPom tryFunc[*jarFile, string] = func(j *jarFile) (string, bool) {
		if j.mvnProject == nil {
			return "", false
		}

		if j.mvnProject.Parent.GroupId == SpringBootStarterGroupId &&
			j.mvnProject.Parent.ArtifactId == SpringBootStarterArtifactId {
			return strings.TrimSpace(j.mvnProject.Parent.Version), len(strings.TrimSpace(j.mvnProject.Parent.Version)) > 0
		}
		return "", false
	}

	var funcs = tryFuncs[*jarFile, string]{tryPom, tryManifest}
	if value, ok := funcs.try(j); ok {
		return value, nil
	}
	return "", nil
}

func (j *jarFile) GetApplicationConfigurations() (map[string]string, error) {
	return j.applicationConfigurations, nil
}

func (j *jarFile) GetLoggingFiles() (map[string]string, error) {
	return j.loggingConfigs, nil
}

func (j *jarFile) GetDependencies() ([]string, error) {
	return j.dependencies, nil
}

func (j *jarFile) GetCertificates() ([]string, error) {
	return j.certificates, nil
}

func (j *jarFile) GetStaticFiles() ([]string, error) {
	return j.staticFiles, nil
}

func (j *jarFile) GetLocation() string {
	return j.remoteLocation
}

func (j *jarFile) GetLastModifiedTime() (time.Time, error) {
	return j.lastModifiedTime, nil
}

func (j *jarFile) GetSize() (int64, error) {
	return j.size, nil
}

func (j *jarFile) GetManifests() map[string]string {
	return j.manifests
}

func (j *jarFile) GetMavenProject() *mvnparser.MavenProject {
	return j.mvnProject
}

func parseManifests(content string) map[string]string {
	scanner := bufio.NewScanner(bytes.NewBufferString(content))
	manifests := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, ":")

		if idx > 0 {
			manifests[strings.TrimSpace(line[:idx])] = strings.TrimSpace(line[idx+1:])
		} else {
			manifests[strings.TrimSpace(line)] = ""
		}
	}
	return manifests
}

func isLoggingConfig(filename string) bool {
	for _, p := range Patterns.LoggingPatterns {
		if match := p.MatchString(filepath.Base(filename)); match {
			return true
		}
	}
	return false
}

func isCertificate(filename string) bool {
	suffix := filepath.Ext(filename)

	return Contains(YamlCfg.Pattern.Cert, suffix)
}

func isStaticContent(filename string) bool {
	for _, folder := range YamlCfg.Pattern.Static.Folder {
		if strings.Contains(filename, folder) {
			return true
		}
	}
	suffix := filepath.Ext(filename)
	if Contains(YamlCfg.Pattern.Static.Extension, suffix) {
		return true
	}

	return false
}

func isAppConfig(filename string) bool {
	for _, p := range Patterns.AppPatterns {
		if match := p.MatchString(filepath.Base(filename)); match {
			return true
		}
	}
	return false
}

func isPomFile(filename string) bool {
	return strings.HasPrefix(filename, DefaultMvnPath) && strings.EqualFold(PomFileName, filepath.Base(filename))
}

func readFileInArchive(f *zip.File) (string, error) {
	var fileInArchive io.ReadCloser
	var err error
	fileInArchive, err = f.Open()
	if err != nil {
		return "", err
	}
	defer fileInArchive.Close()
	var content []byte
	content, err = io.ReadAll(fileInArchive)
	if err != nil {
		return "", errors.Wrap(err, "failed to read file from archive")
	}

	return string(content), nil
}

func readPom(pom string) (*mvnparser.MavenProject, error) {
	var project mvnparser.MavenProject

	if err := xml.Unmarshal([]byte(pom), &project); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal pom")
	}

	return &project, nil
}
