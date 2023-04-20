package springboot

import (
	"archive/zip"
	"path/filepath"
	"strings"
)

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
