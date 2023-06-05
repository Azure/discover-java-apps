package weblogic

import "fmt"

const (
	WeblogicProcessScanCmd = "ps axo pid,uid,cmd|grep weblogic.Server| grep -v grep"

	LinuxSha256Cmd            = "sha256sum %s | awk '{print $1}'"
	LinuxGetJdkVersionCmd     = "%s -version 2>&1 | head -n 1 | awk -F '\"' '{print $2}'"
	LinuxGetTotalMemoryCmd    = "cat /proc/meminfo | grep MemTotal | awk '{print $2}'"
	LinuxGetDefaultMaxHeapCmd = "%s -XX:+PrintFlagsFinal 2>1 | grep ' MaxHeapSize ' | awk '{print $4}'"
	LinuxGetPortsCmd          = `ls -lta /proc/%[1]d/fd | grep socket | awk -F'[\\[\\]]' '{print $2}' | xargs -I {} grep {} /proc/%[1]d/net/tcp /proc/%[1]d/net/tcp6 | awk '{print $3}' | awk -F':' '{print $2}' | sort | uniq | xargs -I {} printf '%%d\n' '0x{}'`
	LinuxGetOsName            = "grep '^ID=' /etc/os-release | awk -F= '{print $2}'"
	LinuxGetOsVersion         = "grep '^VERSION_ID=' /etc/os-release | awk -F= '{print $2}'"
	OracleOsGetName           = "cat /etc/oracle-release | awk '{print $1}'"
	OracleGetVersion          = "cat /etc/oracle-release | awk '{print $3}'"
)

func GetWeblogicProcessScanCmd() string {
	return WeblogicProcessScanCmd
}

func GetSha256Cmd(filename string) string {
	return fmt.Sprintf(LinuxSha256Cmd, filename)
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

func GetOracleOsName() string {
	return fmt.Sprintf(OracleOsGetName)
}

func GetOracleOsVersion() string {
	return fmt.Sprintf(OracleGetVersion)
}
