package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const version = "0.3.0-dev"

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

func addFirewallRule(port string) (ruleName string, created bool) {
	if runtime.GOOS != "windows" {
		return "", false
	}
	if len(port) > 0 && port[0] == ':' { // trim leading ':'
		port = port[1:]
	}
	if !isAdmin() {
		log.Printf("[防火墙] 当前进程非管理员，跳过自动添加。若需外部访问请手动执行: netsh advfirewall firewall add rule name=\"DHGoHttp-%s\" dir=in action=allow protocol=TCP localport=%s", port, port)
		return "", false
	}
	ruleName = "DHGoHttp-" + port
	if firewallRuleExists(ruleName) {
		log.Printf("[防火墙] 规则已存在，跳过创建: %s", ruleName)
		return ruleName, false
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
		"name="+ruleName, "dir=in", "action=allow", "protocol=TCP", "localport="+port)
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("[防火墙] 添加规则失败: %v, 输出: %s", err, string(output))
		return ruleName, false
	}
	log.Printf("[防火墙] 已创建并开放 TCP 端口 %s (规则: %s)", port, ruleName)
	return ruleName, true
}

func main() {
	// Flags
	noFirewall := flag.Bool("no-firewall", false, "跳过自动防火墙规则与提权逻辑")
	elevatedFlag := flag.Bool("elevated", false, "(内部使用) 已提升标记")
	portFlag := flag.Int("port", 0, "指定监听端口 (优先于 PORT 环境变量)")
	maxScanFlag := flag.Int("max-port-scan", 50, "端口占用递增最大尝试次数 (默认 50)")
	dirFlag := flag.String("dir", "", "指定共享根目录 (默认为当前工作目录)")
	bindFlag := flag.String("bind", "", "绑定地址 (默认 0.0.0.0; 为空表示所有网卡)")
	tokenFlag := flag.String("token", "", "访问需要携带的 Token (Header: X-Token 或查询参数 token)")
	readonlyFlag := flag.Bool("readonly", false, "只读模式：禁止目录列出")
	flag.Parse()

	// 获取根目录：优先使用 -dir
	var rootDir string
	if *dirFlag != "" {
		abs, err := filepath.Abs(*dirFlag)
		if err != nil {
			log.Fatalf("[目录] 无法解析路径 %s: %v", *dirFlag, err)
		}
		info, err := os.Stat(abs)
		if err != nil {
			log.Fatalf("[目录] 目录不存在或无法访问: %s (%v)", abs, err)
		}
		if !info.IsDir() {
			log.Fatalf("[目录] 指定路径不是目录: %s", abs)
		}
		rootDir = abs
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("无法获取当前目录:", err)
		}
		rootDir = cwd
	}

	// 解析端口：优先级 flag > 环境变量 PORT > 默认 8080
	basePort := 8080
	if *portFlag > 0 && *portFlag < 65536 {
		basePort = *portFlag
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			basePort = p
		} else if envPort != "" {
			log.Printf("[端口] 无效的 PORT 环境变量值: %s，已回退使用 %d", envPort, basePort)
		}
	}

	finalPort, tried, err := findAvailablePort(basePort, *maxScanFlag)
	if err != nil {
		log.Fatalf("[端口] 无法找到可用端口，起始 %d: %v", basePort, err)
	}

	if len(tried) > 1 {
		log.Printf("[端口] 基础端口 %d 被占用或不可用，尝试序列: %v -> 选用 %d", basePort, tried, finalPort)
	} else {
		log.Printf("[端口] 使用端口: %d", finalPort)
	}

	log.Printf("DHGoHttp v%s 启动，根目录: %s", version, rootDir)
	log.Printf("示例: curl http://localhost:%d/your-file.sh", finalPort)

	// 创建处理链
	fileServer := http.FileServer(http.Dir(rootDir))
	var handler http.Handler = http.StripPrefix("/", fileServer)
	if *readonlyFlag {
		handler = readonlyMiddleware(handler, rootDir)
	}
	if *tokenFlag != "" {
		handler = tokenMiddleware(handler, *tokenFlag)
	}
	handler = loggingMiddleware(handler)
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	// 启动服务器：使用最终确定端口与绑定地址
	addrHost := ""
	if *bindFlag != "" {
		addrHost = *bindFlag
	}
	listenAddr := fmt.Sprintf("%s:%d", addrHost, finalPort)

	var createdRule string
	var created bool
	if *noFirewall {
		log.Printf("[启动] 检测到 -no-firewall，跳过提权与防火墙规则创建。")
	} else {
		if runtime.GOOS == "windows" && !isAdmin() && !*elevatedFlag {
			if tryElevate() {
				log.Printf("已请求以管理员权限重新启动，当前进程退出。")
				return
			}
			log.Printf("[权限] 用户可能取消了提权，继续以普通权限运行。")
		}
		if runtime.GOOS == "windows" {
			if rule, c := addFirewallRule(fmt.Sprintf(":%d", finalPort)); rule != "" {
				createdRule, created = rule, c
			}
		}
	}
	log.Printf("[启动] 服务正在监听: http://%s/", displayURL(*bindFlag, finalPort))
	if *tokenFlag != "" {
		log.Printf("[安全] 已启用 Token 校验 (X-Token / ?token=)")
	}
	if *readonlyFlag {
		log.Printf("[模式] 已启用只读目录（禁目录列出）")
	}

	srv := &http.Server{Addr: listenAddr, Handler: mux}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Printf("[退出] 收到信号，正在关闭...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("[退出] 关闭 HTTP 服务出错: %v", err)
		}
		if createdRule != "" && created {
			if err := deleteFirewallRule(createdRule); err != nil {
				log.Printf("[防火墙] 删除规则失败(%s): %v", createdRule, err)
			} else {
				log.Printf("[防火墙] 已删除规则: %s", createdRule)
			}
		}
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("服务器错误: %v", err)
	}
	log.Printf("[退出] 服务器已关闭。")
}

// tryElevate attempts to relaunch the current executable with administrative privileges (UAC prompt).
// Returns true if an elevation attempt was started (regardless of user approval), false if not attempted or failed early.
func tryElevate() bool {
	exe, err := os.Executable()
	if err != nil {
		log.Printf("[权限] 获取可执行文件路径失败: %v", err)
		return false
	}
	// 构造参数：保留原有参数，追加 -elevated，若未显式传 -dir 则追加当前绝对目录
	argsList := os.Args[1:] // 原始除去程序本身
	hasDir := false
	for _, a := range argsList {
		if strings.HasPrefix(a, "-dir") { // -dir 或 -dir=...
			hasDir = true
			break
		}
	}
	// 追加 -elevated（若未有）
	needElevated := true
	for _, a := range argsList {
		if a == "-elevated" || strings.HasPrefix(a, "-elevated=") {
			needElevated = false
			break
		}
	}
	if needElevated {
		argsList = append(argsList, "-elevated")
	}
	if !hasDir {
		cwd, err := os.Getwd()
		if err == nil {
			abs, _ := filepath.Abs(cwd)
			argsList = append(argsList, "-dir", abs)
		}
	}
	// 拼接为单一字符串（ShellExecuteW 参数）
	combined := strings.Join(argsList, " ")

	verbPtr, _ := syscall.UTF16PtrFromString("runas")
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwd := filepath.Dir(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argsPtr, _ := syscall.UTF16PtrFromString(combined)

	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	r, _, err := proc.Call(0,
		uintptr(unsafe.Pointer(verbPtr)),
		uintptr(unsafe.Pointer(exePtr)),
		uintptr(unsafe.Pointer(argsPtr)),
		uintptr(unsafe.Pointer(cwdPtr)),
		uintptr(1))
	if r <= 32 { // 失败
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

// deleteFirewallRule removes a firewall rule by name (Windows only, ignore if missing)
func deleteFirewallRule(ruleName string) error {
	if runtime.GOOS != "windows" || ruleName == "" {
		return nil
	}
	cmd := exec.Command("netsh", "advfirewall", "firewall", "delete", "rule", "name="+ruleName)
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "No rules match") || strings.Contains(string(output), "找不到与指定条件匹配的规则") {
			return nil
		}
		return fmt.Errorf("delete rule failed: %s err=%w", string(output), err)
	}
	return nil
}

// logging middleware and helpers
type logResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (l *logResponseWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

func (l *logResponseWriter) Write(b []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	n, err := l.ResponseWriter.Write(b)
	l.size += n
	return n, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &logResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		dur := time.Since(start)
		ip := clientIP(r)
		log.Printf("[访问] %s %s %s %d %dB %s", ip, r.Method, r.URL.Path, lrw.status, lrw.size, dur)
	})
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func tokenMiddleware(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		provided := r.Header.Get("X-Token")
		if provided == "" {
			provided = r.URL.Query().Get("token")
		}
		if provided != token {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, "Unauthorized\n")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func readonlyMiddleware(next http.Handler, root string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if strings.HasSuffix(p, "/") || p == "" { // directory style request
			abs := filepath.Join(root, filepath.FromSlash(p))
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				w.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(w, "Directory listing disabled\n")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func displayURL(bind string, port int) string {
	h := bind
	if h == "" || h == "0.0.0.0" || h == "::" {
		h = "localhost"
	}
	return fmt.Sprintf("%s:%d", h, port)
}

// findAvailablePort tries to listen starting from startPort, incrementing by 1 until a free port is found
// or attempts exceed maxAttempts. It returns the selected port, the slice of tried ports (in order),
// and an error if none were available. The listener is immediately closed after a successful probe.
func findAvailablePort(startPort int, maxAttempts int) (int, []int, error) {
	tried := make([]int, 0, maxAttempts)
	for i := 0; i < maxAttempts; i++ {
		p := startPort + i
		tried = append(tried, p)
		if p <= 0 || p >= 65536 {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err != nil {
			continue // occupied or not permitted
		}
		// success
		_ = ln.Close()
		return p, tried, nil
	}
	return 0, tried, fmt.Errorf("no free port found in range %d-%d", startPort, startPort+maxAttempts-1)
}
