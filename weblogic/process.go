package weblogic

import (
	"bufio"
	"bytes"
	"crypto/rand"
	_ "embed"
	"fmt"
	"github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"io"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	JavaCmd                   = "java"
	JarOption                 = "-jar"
	JvmOptionXmx              = "-Xmx"
	JvmOptionMaxRamPercentage = "-XX:MaxRAMPercentage"
	KiB                       = 1024
	MiB                       = KiB * 1024
	WeblogicName              = "-Dweblogic.Name="
	WeblogicHome              = "-Dweblogic.home="
	LetterBytes               = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

//go:embed weblogic-deploy.zip
var wdtFileData []byte

//go:embed discover_wls_domain_home.py
var discoverDomainHome []byte

//go:embed discover_absolute_path.py
var discoverAbsolutePath []byte

type javaProcess struct {
	pid          int
	uid          int
	options      []string
	environments []string
	javaCmd      string
	executor     ServerDiscovery
}

func (p *javaProcess) GetLastModifiedTime(path string) (time.Time, error) {
	println("GetLastModifiedTime of application  ....")

	commands := []string{
		"file=" + path,
		"stat -c %Y $file",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}

	timestamp, _ := strconv.ParseInt(strings.TrimSpace(output), 10, 64)

	println("last modified time: " + time.Unix(timestamp, 0).String())
	return time.Unix(timestamp, 0), nil
}

func (p *javaProcess) GetApplicationsAndPath(DomainHome string, Randomfolder string, connection string) map[string]string {
	println("Getting application names and app paths ......")

	var sshClient = p.executor.Server().Client()
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()

	remoteFilePath := Randomfolder + "/discover_absolute_path.py"

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		panic(err)
	}
	defer remoteFile.Close()

	buf := bytes.NewBuffer(discoverAbsolutePath)
	_, err = io.Copy(remoteFile, buf)
	if err != nil {
		panic(err)
	}

	filePath := Randomfolder + "/discover_absolute_path.py"
	commands := []string{
		"echo \"" + connection + "\" | cat - " + filePath + " > temp && mv temp " + filePath,
		". " + DomainHome + "/bin/setDomainEnv.sh; java $WLST_ARGS weblogic.WLST " + filePath,
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}

	return extractValues(output)
}

func extractValues(input string) map[string]string {
	regex := regexp.MustCompile(`application_name is: (.*?); absolute_path is: (.*?);`)
	matches := regex.FindAllStringSubmatch(input, -1)

	result := make(map[string]string)
	for _, match := range matches {
		applicationName := match[1]
		absolutePath := match[2]
		result[applicationName] = absolutePath
	}

	return result
}

func (p *javaProcess) DeleteTempFolder(path string) string {

	println("Deleting RandomFolder ....")

	commands := []string{
		"rm -rf " + path + "/",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}
	return strings.Trim(output, " \n")
}

func (p *javaProcess) CreateTempFolder() string {
	randomString := generateRandomString(5)
	randomFolder := "discover_weblogic_" + randomString

	commands := []string{
		"mkdir " + randomFolder,
		"cd " + randomFolder,
		"pwd",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}
	return strings.Trim(output, " \n")
}

func (p *javaProcess) GetDomainHome(Oracle_Home string, Randomfolder string, connection string) (string, error) {
	println("Getting Domain_Home ......")

	var sshClient = p.executor.Server().Client()
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()

	remoteFilePath := Randomfolder + "/discover_wls_domain_home.py"

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		panic(err)
	}
	defer remoteFile.Close()

	buf := bytes.NewBuffer(discoverDomainHome)
	_, err = io.Copy(remoteFile, buf)
	if err != nil {
		panic(err)
	}

	commands := []string{
		"cd " + Randomfolder,
		"export Oracle_Home=" + Oracle_Home,
		"echo \"" + connection + "\" | cat - " + remoteFilePath + " > temp && mv temp " + remoteFilePath,
		"bash $Oracle_Home/oracle_common/common/bin/wlst.sh " + remoteFilePath,
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}
	return extractDomainHome(output)
}

func (p *javaProcess) GetWeblogicPatch(DomainHome string) string {
	println("Getting Weblogic Patch ......")

	commands := []string{
		". " + DomainHome + "/bin/setDomainEnv.sh",
		"java weblogic.version -verbose",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}

	return findOPatchPatches(output)
}

func (p *javaProcess) GetWeblogicVersion(DomainHome string) string {
	println("Getting Weblogic Version ......")

	commands := []string{
		". " + DomainHome + "/bin/setDomainEnv.sh",
		"java weblogic.version -verbose",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}

	return findWebLogicVersion(output)
}

func (p *javaProcess) GetDiscoverDomainResult(Randomfolder string) string {
	println("Getting discoverDomain.sh runing result....start")

	command := "cat " + Randomfolder + "/wlsdModel.yaml"
	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}
	println("output is: ", output)
	return output
}

func (p *javaProcess) RunDiscoverDomainCommand(randomfolder, javaHome, weblogicUser, weblogicPassword, oracleHome string, port int) string {

	println("Running RunDiscoverDomainCommand in WDT....")

	commands := []string{
		"export PSW=" + weblogicPassword,
		"export JAVA_HOME=" + javaHome,
		randomfolder + "/weblogic-deploy/bin/discoverDomain.sh " +
			"-oracle_home " + oracleHome + " " +
			"-remote  " +
			"-model_file  " + randomfolder + "/wlsdModel.yaml  " +
			"-admin_user " + weblogicUser + " " +
			"-admin_url t3://localhost:" + strconv.Itoa(port) + " " +
			"-admin_pass_env PSW",
	}
	command := strings.Join(commands, "; ")

	output, err := p.executor.Server().RunCmd(command)
	if err != nil {
		panic(err)
	}

	return output

}

func (p *javaProcess) UploadAndInstallWDT(Randomfolder string) string {

	var sshClient = p.executor.Server().Client()
	println("Uploading WebLogic Deploy Tooling (WDT) to the remote server ......")
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		panic(err)
	}
	defer sftpClient.Close()

	remoteFilePath := Randomfolder + "/weblogic-deploy.zip" // Replace with the remote file path where you want to upload

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		panic(err)
	}
	defer remoteFile.Close()

	buf := bytes.NewBuffer(wdtFileData)
	_, err = io.Copy(remoteFile, buf)
	if err != nil {
		panic(err)
	}

	// Create a new SSH session to run the unzip command
	session, err := sshClient.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	commands := []string{
		"cd " + Randomfolder,
		"unzip " + remoteFilePath,
	}
	command := strings.Join(commands, "; ")

	if err := session.Run(command); err != nil {
		panic(err)
	}

	return "success upload wdt"
}

func (p *javaProcess) GetWeblogicName() (string, error) {

	var result string
	for _, option := range p.options {
		if strings.HasPrefix(option, WeblogicName) {
			return strings.TrimPrefix(option, WeblogicName), nil
		}
	}
	return result, nil
}

func (p *javaProcess) GetWeblogicHome() (string, error) {
	var result string
	for _, option := range p.options {
		if strings.HasPrefix(option, WeblogicHome) {
			return strings.TrimPrefix(option, WeblogicHome), nil
		}
	}
	return result, nil
}

func (p *javaProcess) GetJavaHome() (string, error) {
	return strings.TrimSuffix(p.javaCmd, "/bin/java"), nil
}

func (p *javaProcess) GetRuntimeJdkVersion() (string, error) {
	buf, err := runWithSudo(p.executor.Server(), GetJdkVersionCmd(p.javaCmd))
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
	output, err := runWithSudo(p.executor.Server(), GetDefaultMaxHeap(p.javaCmd))
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

func (p *javaProcess) GetUid() int {
	return p.uid
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

func findWebLogicVersion(str string) string {
	index := strings.Index(str, "WebLogic Server ")
	if index != -1 {
		startIndex := index + len("WebLogic Server ")
		endIndex := strings.Index(str[startIndex:], " ")
		if endIndex != -1 {
			return str[startIndex : startIndex+endIndex]
		}
	}
	return ""
}

func findOPatchPatches(str string) string {
	index := strings.Index(str, "OPatch Patches:")
	if index != -1 {
		startIndex := index + len("OPatch Patches:\n")
		endIndex := strings.Index(str[startIndex:], "\nSERVICE NAME")
		if endIndex != -1 {
			return str[startIndex : startIndex+endIndex]
		}
	}
	return ""
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(LetterBytes))))
		b[i] = LetterBytes[num.Int64()]
	}
	return string(b)
}

func extractDomainHome(text string) (string, error) {
	// Define the regular expression pattern to match the domain_home value
	pattern := `The domain_home is:\s+(.+)`

	// Compile the regular expression
	regex := regexp.MustCompile(pattern)

	// Find the first match
	match := regex.FindStringSubmatch(text)

	if len(match) > 1 {
		return match[1], nil
	} else {
		return "", fmt.Errorf("No domain_home found")
	}
}
