package springboot

import (
	"bufio"
	"fmt"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"math"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	JavaCmd                   = "java"
	JarOption                 = "-jar"
	JvmOptionXmx              = "-Xmx"
	JvmOptionMaxRamPercentage = "-XX:MaxRAMPercentage"
	KiB                       = 1024
	MiB                       = KiB * 1024
)

type javaProcess struct {
	pid          int
	uid          int
	options      []string
	environments []string
	javaCmd      string
	executor     ServerDiscovery
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
		return "", errors.New(fmt.Sprintf("jar file not found in process %d", p.pid))
	}
	if !filepath.IsAbs(jarFileName) {
		// when jar file path is not absolute path, we shall locate the jar file path again
		output, err := p.executor.Server().RunCmd(GetLocateJarCmd(p.pid, filepath.Base(jarFileName)))
		if err != nil {
			return "", err
		}
		if len(output) == 0 {
			return "", errors.New(fmt.Sprintf("cannot locate jar: %s", jarFileName))
		}
		absolutePath = output
	} else {
		absolutePath = jarFileName
	}

	return CleanOutput(absolutePath), nil
}

func (p *javaProcess) GetRuntimeJdkVersion() (string, error) {
	buf, err := p.executor.Server().RunCmd(GetJdkVersionCmd(p.javaCmd))
	if err != nil {
		return "", err
	}

	return CleanOutput(buf), nil
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
				if !Contains(YamlCfg.Env.Denylist, envName) {
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

func (p *javaProcess) GetJavaCmd() (string, error) {
	return p.javaCmd, nil
}

func (p *javaProcess) GetJvmMemory() (int64, error) {
	for _, option := range p.options {
		if strings.HasPrefix(option, JvmOptionXmx) {
			bs, err := units.RAMInBytes(option[len(JvmOptionXmx):])
			if err != nil {
				return 0, errors.Wrap(err, fmt.Sprintf("failed to parse -Xmx from pid %v", p.pid))
			}
			return bs, nil
		}
	}

	// Do a second iteration here due to -Xmx has higher priority than -XX:MaxRamPercentage
	// If both are set, -XX:MaxRamPercentage will be ignored
	// So if nothing found in the first iteration, we try another round
	for _, option := range p.options {
		if strings.HasPrefix(option, JvmOptionMaxRamPercentage) {
			total, err := p.executor.GetTotalMemory()
			if err != nil {
				return 0, err
			}

			percent, err := strconv.ParseFloat(option[len(JvmOptionMaxRamPercentage)+1:], 64)
			if err != nil {
				return 0, errors.Wrap(err, "failed to parse -XX:MaxRAMPercentage")
			}
			return int64(math.Round(float64(total)*percent) / 100), nil
		}
	}

	defaultMaxHeap, err := p.getDefaultMaxHeapSize()
	if err != nil {
		return 0, err
	}
	return defaultMaxHeap, nil
}

func (p *javaProcess) Executor() ServerDiscovery {
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

func (p *javaProcess) getDefaultMaxHeapSize() (int64, error) {
	output, err := p.Executor().Server().RunCmd(GetDefaultMaxHeap(p.javaCmd))
	if err != nil {
		return 0, err
	}
	if len(output) == 0 {
		return 0, errors.New("failed to get default MaxHeapSize, output is empty")
	}

	size, err := strconv.ParseFloat(CleanOutput(output), 64)
	if err != nil {
		return 0, errors.Wrap(err, fmt.Sprintf("failed to parse default MaxHeapSize, output: %s", output))
	}

	return int64(size), nil
}

func (p *javaProcess) String() string {
	s := []string{strconv.Itoa(p.pid)}
	s = append(s, p.options...)
	return strings.Join(s, " ")
}

func runWithSudo(server ServerConnector, cmd string) (string, error) {
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
