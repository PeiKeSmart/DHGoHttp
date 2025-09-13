package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// addFirewallRule attempts to add a Windows Defender Firewall inbound rule for the given TCP port.
// It is a no-op on non-Windows platforms. If the rule already exists, it will ignore the error.
// isAdmin returns whether current process has administrative privileges on Windows.
// For non-Windows it always returns false.
func isAdmin() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	// Build Administrators group SID
	sid, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return false
	}
	// Check token membership
	token := windows.Token(0)
	member, err := token.IsMember(sid)
	if err != nil {
		return false
	}
	return member
}

func addFirewallRule(port string) {
	if runtime.GOOS != "windows" {
		return
	}
	// Remove leading colon if passed like ":8080"
	if len(port) > 0 && port[0] == ':' {
		port = port[1:]
	}

	if !isAdmin() {
		log.Printf("[防火墙] 当前进程非管理员，自动添加防火墙规则将跳过。若需外部访问，请以管理员权限运行或手动执行: netsh advfirewall firewall add rule name=\"DHGoHttp-%s\" dir=in action=allow protocol=TCP localport=%s", port, port)
		return
	}
	// Use netsh to add a firewall rule. This requires administrator privileges.
	// Before adding, check if rule exists.
	ruleName := "DHGoHttp-" + port
	if firewallRuleExists(ruleName) {
		log.Printf("[防火墙] 规则已存在，跳过创建: %s", ruleName)
		return
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name="+ruleName, "dir=in", "action=allow", "protocol=TCP", "localport="+port)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[防火墙] 添加规则失败: %v, 输出: %s", err, string(output))
	} else {
		log.Printf("[防火墙] 已创建并开放 TCP 端口 %s (规则: %s)", port, ruleName)
	}
}

func main() {
	// Flags
	noFirewall := flag.Bool("no-firewall", false, "跳过自动防火墙规则与提权逻辑")
	elevatedFlag := flag.Bool("elevated", false, "(内部使用) 已提升标记")
	flag.Parse()

	// 获取当前工作目录作为根目录
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法获取当前目录:", err)
	}

	log.Printf("文件服务器启动，根目录: %s", rootDir)
	log.Printf("服务器运行在: http://localhost:8080")
	log.Printf("使用 curl http://localhost:8080/your-file.sh 下载文件")

	// 创建文件服务器处理器
	fileServer := http.FileServer(http.Dir(rootDir))

	// 设置路由
	http.Handle("/", http.StripPrefix("/", fileServer))

	// 启动服务器
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	if *noFirewall {
		log.Printf("[启动] 检测到 -no-firewall，跳过提权与防火墙规则创建。")
	} else {
		// Windows 下尝试提权（仅在非管理员且未带 -elevated 标记）
		if runtime.GOOS == "windows" && !isAdmin() && !*elevatedFlag {
			if tryElevate() {
				log.Printf("已请求以管理员权限重新启动，当前进程退出。")
				return
			}
			log.Printf("[权限] 用户可能取消了提权，继续以普通权限运行。")
		}
		// 尝试开放 Windows 防火墙端口（非 Windows 忽略）
		addFirewallRule(port)
	}

	log.Fatal(http.ListenAndServe(port, nil))
}

// tryElevate attempts to relaunch the current executable with administrative privileges (UAC prompt).
// Returns true if an elevation attempt was started (regardless of user approval), false if not attempted or failed early.
func tryElevate() bool {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[权限] 获取可执行文件路径失败: %v", err)
		return false
	}
	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	wd := filepath.Dir(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(wd)
	// 传递一个标记参数避免递归无限提升
	args := "-elevated"
	argsPtr, _ := syscall.UTF16PtrFromString(args)

	// Use ShellExecuteW (via windows.ShellExecute) but x/sys/windows does not wrap it, use syscall.Syscall6 on shell32.dll.
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	// HINSTANCE ShellExecuteW(HWND hwnd, LPCWSTR lpOperation, LPCWSTR lpFile, LPCWSTR lpParameters, LPCWSTR lpDirectory, INT nShowCmd);
	r, _, err := proc.Call(0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		uintptr(1)) // SW_SHOWNORMAL
	// According to docs, return value > 32 indicates success.
	if r <= 32 {
		log.Printf("[权限] 触发 UAC 提升失败: %v (返回 %d)", err, r)
		return false
	}
	return true
}

// firewallRuleExists checks whether a firewall rule with the given name exists (Windows only).
func firewallRuleExists(ruleName string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "show", "rule", "name="+ruleName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// netsh returns "No rules match the specified criteria." with exit code 1 (on some versions) when not found.
		if strings.Contains(string(output), "No rules match") || strings.Contains(string(output), "找不到与指定条件匹配的规则") {
			return false
		}
		// Other errors: assume not exist but log once (optional). For now just treat as not exist.
		return false
	}
	// Heuristic: if output contains "Enabled:" we assume it's a valid rule dump.
	return strings.Contains(string(output), "Enabled:")
}
