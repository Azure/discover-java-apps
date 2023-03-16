package core

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/go-units"
	"k8s.io/utils/strings/slices"
	"microsoft.com/azure-spring-discovery/api/v1alpha1"
)

const (
	EnvBlackListConfigKey = "config.env.blacklist"
	JavaCmd               = "java"
	JarOption             = "-jar"
)

type JavaProcess interface {
	GetProcessId() int
	GetRuntimeJdkVersion() (string, error)
	LocateJarFile() (string, error)
	GetJvmOptions() ([]string, error)
	GetEnvironments() ([]string, error)
	GetJvmMemoryInMb() (float64, error)
	GetPorts() ([]int, error)
	Executor() DiscoveryExecutor
}

type javaProcess struct {
	pid          int
	uid          int
	options      []string
	environments []string
	javaCmd      string
	executor     DiscoveryExecutor
}

func (p *javaProcess) LocateJarFile() (string, error) {
	var jarFileName string
	var absolutePath string
	for idx, option := range p.options {
		if option == JarOption && idx+1 < len(p.options) {
			jarFileName = p.options[idx+1]
		}
	}

	if len(jarFileName) == 0 {
		return "", DiscoveryError{message: fmt.Sprintf("jar file not found in process %d", p.pid), severity: v1alpha1.Error}
	}
	if !filepath.IsAbs(jarFileName) {
		// when jar file path is not absolute path, we shall locate the jar file path again
		output, err := p.executor.Server().RunCmd(GetLocateJarCmd(p.pid, filepath.Base(jarFileName)))
		if err != nil {
			return "", err
		}
		if len(output) == 0 {
			return "", DiscoveryError{message: fmt.Sprintf("cannot locate jar file in filesystem %s", jarFileName), severity: v1alpha1.Error}
		}
		absolutePath = output
	} else {
		absolutePath = jarFileName
	}

	return cleanOutput(absolutePath, LinuxNewLineCharacter), nil
}

func (p *javaProcess) GetRuntimeJdkVersion() (string, error) {
	buf, err := p.executor.Server().RunCmd(GetJdkVersionCmd(p.javaCmd))
	if err != nil {
		return "", err
	}

	jdkVersion := cleanOutput(buf, LinuxNewLineCharacter)
	return sanitizeVersion(jdkVersion), nil
}

func (p *javaProcess) GetJvmOptions() ([]string, error) {
	var jvmOptions []string
	var jarOpIdx = -1
	for idx, option := range p.options {
		if strings.EqualFold(option, JarOption) {
			jarOpIdx = idx
			continue
		}
		if jarOpIdx != -1 && idx == jarOpIdx+1 {
			// this is jar file
			continue
		}
		jvmOptions = append(jvmOptions, option)
	}

	return jvmOptions, nil
}

var envSplitter bufio.SplitFunc = func(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	// Find the index of the input of a newline followed by a
	// pound sign.
	if i := strings.Index(string(data), "\000"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	// If at end of file with data return the data
	if atEOF {
		return len(data), data, nil
	}

	return
}

func (p *javaProcess) GetEnvironments() ([]string, error) {
	if p.environments == nil {
		output, err := runWithSudo(p.executor.Server(), GetEnvCmd(p.pid))
		if err != nil {
			return nil, err
		}
		var environments []string
		if len(output) == 0 {
			return environments, nil
		}
		scanner := bufio.NewScanner(strings.NewReader(output))
		scanner.Split(envSplitter)

		for scanner.Scan() {
			env := scanner.Text()
			idx := strings.Index(env, "=")

			if idx > 0 {
				envName := env[:idx]
				if !slices.Contains(config.GetConfigListByKey(EnvBlackListConfigKey), envName) {
					environments = append(environments, env)
				}
			} else {
				environments = append(environments, env)
			}
		}
		return environments, nil
	} else {
		return p.environments, nil
	}
}

func (p *javaProcess) GetJvmMemoryInMb() (float64, error) {
	for _, option := range p.options {
		if strings.HasPrefix(option, JvmOptionXmx) {
			bs, err := units.RAMInBytes(option[len(JvmOptionXmx):])
			if err != nil {
				return 0, DiscoveryError{message: fmt.Sprintf("unable to parse -Xmx, pid=%d", p.pid), severity: v1alpha1.Warning}
			}
			return math.Round(float64(bs)*100/MiB) / 100, nil
		}
	}

	// Do a second iteration here due to -Xmx has higher priority than -XX:MaxRamPercentage
	// If both are set, -XX:MaxRamPercentage will be ignored
	// So if nothing found in the first iteration, we try another round
	for _, option := range p.options {
		if strings.HasPrefix(option, JvmOptionMaxRamPercentage) {
			totalInKB, err := p.executor.GetTotalMemoryInKB()
			if err != nil {
				return 0, err
			}

			percent, err := strconv.ParseFloat(option[len(JvmOptionMaxRamPercentage)+1:], 2)
			if err != nil {
				return 0, DiscoveryError{message: fmt.Sprintf("failed to parse -XX:MaxRAMPercentage"), severity: v1alpha1.Warning}
			}

			return math.Round(totalInKB*percent/KiB) / 100, nil
		}
	}

	defaultMaxHeap, err := p.executor.GetDefaultMaxHeapSizeInBytes()
	if err != nil {
		return 0, err
	}
	return math.Round(defaultMaxHeap*100/MiB) / 100, nil
}

func (p *javaProcess) Executor() DiscoveryExecutor {
	return p.executor
}

func (p *javaProcess) GetProcessId() int {
	return p.pid
}

func (p *javaProcess) GetPorts() ([]int, error) {
	output, err := runWithSudo(p.executor.Server(), GetPortsCmd(p.pid))
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(strings.NewReader(output))
	var ports []int
	for scanner.Scan() {
		text := scanner.Text()
		if len(strings.TrimSpace(text)) > 0 {
			if port, e := strconv.Atoi(strings.TrimSpace(text)); e == nil {
				ports = append(ports, port)
			}
		}
	}

	return ports, nil
}

func (p *javaProcess) String() string {
	s := []string{strconv.Itoa(p.pid)}
	s = append(s, p.options...)
	return strings.Join(s, " ")
}

func runWithSudo(server TargetServer, cmd string) (string, error) {
	output, err := server.RunCmd(cmd)
	if err != nil {
		if errors.As(err, &PermissionDenied{}) {
			output, err = server.RunCmd(sudo(cmd))
		}
		if err != nil {
			return "", err
		}
	}
	return output, nil
}
