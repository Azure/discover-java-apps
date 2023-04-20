package springboot

import "fmt"

const (
	LinuxProcessScanCmd       = "ps axo pid,uid,cmd | grep [j]ava | grep '\\-jar' | grep -v grep"
	LinuxLocateJarCmd         = "ls -l /proc/%d/fd | grep %s | head -1 | awk '{print $11}'"
	LinuxSha256Cmd            = "sha256sum %s | awk '{print $1}'"
	LinuxGetEnvCmd            = "cat /proc/%d/environ"
	LinuxGetJdkVersionCmd     = "%s -version 2>&1 | head -n 1 | awk -F '\"' '{print $2}'"
	LinuxGetTotalMemoryCmd    = "cat /proc/meminfo | grep MemTotal | awk '{print $2}'"
	LinuxGetDefaultMaxHeapCmd = "%s -XX:+PrintFlagsFinal 2>1 | grep ' MaxHeapSize ' | awk '{print $4}'"
	LinuxGetPortsCmd          = `ls -lta /proc/%[1]d/fd | grep socket | awk -F'[\\[\\]]' '{print $2}' | xargs -I {} grep {} /proc/%[1]d/net/tcp /proc/%[1]d/net/tcp6 | awk '{print $3}' | awk -F':' '{print $2}' | sort | uniq | xargs -I {} printf '%%d\n' '0x{}'`
	LinuxGetOsName            = "grep '^ID=' /etc/os-release | awk -F= '{print $2}'"
	LinuxGetOsVersion         = "grep '^VERSION_ID=' /etc/os-release | awk -F= '{print $2}'"
	CentOsGetName             = "cat /etc/centos-release | awk '{print $1}'"
	CentOsGetVersion          = "cat /etc/centos-release | awk '{print $3}'"
)

func GetProcessScanCmd() string {
	return LinuxProcessScanCmd
}

func GetLocateJarCmd(pid int, filename string) string {
	return fmt.Sprintf(LinuxLocateJarCmd, pid, filename)
}

func GetSha256Cmd(filename string) string {
	return fmt.Sprintf(LinuxSha256Cmd, filename)
}

func GetEnvCmd(pid int) string {
	return fmt.Sprintf(LinuxGetEnvCmd, pid)
}

func GetJdkVersionCmd(javaCmd string) string {
	return fmt.Sprintf(LinuxGetJdkVersionCmd, javaCmd)
}

func GetTotalMemoryCmd() string {
	return LinuxGetTotalMemoryCmd
}

func GetDefaultMaxHeap(javaCmd string) string {
	return fmt.Sprintf(LinuxGetDefaultMaxHeapCmd, javaCmd)
}

func GetPortsCmd(pid int) string {
	return fmt.Sprintf(LinuxGetPortsCmd, pid)
}

func GetOsName() string {
	return fmt.Sprintf(LinuxGetOsName)
}

func GetOsVersion() string {
	return fmt.Sprintf(LinuxGetOsVersion)
}

func GetCentOsName() string {
	return fmt.Sprintf(CentOsGetName)
}

func GetCentOsVersion() string {
	return fmt.Sprintf(CentOsGetVersion)
}
