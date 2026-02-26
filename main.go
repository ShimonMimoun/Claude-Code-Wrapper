package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	prompt "github.com/c-bata/go-prompt"
)

// --- Infrastructure Configuration ---
const (
	SsoBaseURL      = "https://ai-proxy.domain.local/sso/login"
	ProxyBaseURL    = "https://ai-proxy.domain.local/v1"
	ConfigAPIURL    = "https://ai-proxy.domain.local/api/claude-settings"
	CliDownloadBase = "https://ai-proxy.domain.local" // Server root for binaries
	DefaultPort     = 8080
	MaxPortRetries  = 10
)

func ssoURL(port int) string {
	return fmt.Sprintf("%s?redirect=http://127.0.0.1:%d/callback", SsoBaseURL, port)
}

// ANSI colors inspired by the Wrapper logo
const (
	ColorRed    = "\033[38;2;240;0;11m"
	ColorDark   = "\033[1;30m"
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

func main() {
	printLogo(false)

	// 1. Initial authentication check
	doneSpinner := make(chan bool)
	go showSpinner("Verifying Wrapper identity...", doneSpinner)
	token := getToken()

	time.Sleep(1 * time.Second) // Decorative delay for the animation
	doneSpinner <- true

	if token == "" || !isTokenValid(token) {
		fmt.Printf("\n%sAuthentication required. Launching Wrapper SSO...%s\n", ColorYellow, ColorReset)
		token = authenticateViaSSO()
		if token != "" {
			saveToken(token)
			fmt.Printf("\n%s✅ Wrapper connected successfully!%s\n\n", ColorGreen, ColorReset)
		} else {
			fmt.Printf("\n%s❌ Authentication failed. Dropping into Wrapper shell — use /login to retry.%s\n", ColorRed, ColorReset)
			WrapperShell()
			os.Exit(0)
		}
	} else {
		fmt.Printf("\n%s✅ Wrapper session active.%s\n\n", ColorGreen, ColorReset)
	}

	// 2. Fetch and merge enterprise settings
	fetchAndMergeSettings(token)

	// 3. Verify the presence of the native Claude executable
	ensureCliExists(token)

	// 4. Start background token check (every 10 minutes)
	go startBackgroundTokenCheck()

	fmt.Printf("\n%s💡 TIP: Keep this terminal open! Wrapper will automatically verify your SSO token every 10 minutes.%s\n\n", ColorYellow, ColorReset)

	// 5. MAIN LOOP: Launch Claude, then drop into Wrapper Shell when it exits
	for {
		launchClaudeCode()
		WrapperShell()
	}
}

// --- Wrapper Interactive Shell ---

// shellSuggestions returns command suggestions for go-prompt
func shellSuggestions(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "/claude", Description: "Open Claude Code"},
		{Text: "/restart", Description: "Launch Claude Code again"},
		{Text: "/start", Description: "Launch Claude Code again"},
		{Text: "/login", Description: "Force a new SSO authentication"},
		{Text: "/reload", Description: "Refresh enterprise settings"},
		{Text: "/refresh", Description: "Refresh enterprise settings"},
		{Text: "/sync-models", Description: "Fetch latest models from proxy"},
		{Text: "/usage", Description: "View usage statistics"},
		{Text: "/about", Description: "Show version and info"},
		{Text: "/help", Description: "List all commands"},
		{Text: "/exit", Description: "Close Wrapper"},
		{Text: "/quit", Description: "Close Wrapper"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func WrapperShell() {
	fmt.Printf("\n%s──────────────────────────────────────────────────────────────────────────────────────────────────%s\n", ColorDark, ColorReset)
	
	isLoggedIn := getToken() != "" && isTokenValid(getToken())
	printLogo(isLoggedIn)
	fmt.Printf("\n%s[Wrapper] Claude Code session completed.%s\n", ColorDark, ColorReset)
	fmt.Printf("\n%s💡 Tip: Do not close this window to keep your automatic token renewal active.%s\n", ColorYellow, ColorReset)
	fmt.Printf("\n%sType /help to see all available commands.%s\n\n\n", ColorDark, ColorReset)

	for {
		input := prompt.Input(
			"Wrapper> ",
			shellSuggestions,
			prompt.OptionPrefixTextColor(prompt.Red),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
			prompt.OptionSelectedSuggestionBGColor(prompt.Red),
			prompt.OptionDescriptionBGColor(prompt.DarkGray),
			prompt.OptionSelectedDescriptionBGColor(prompt.Red),
		)

		input = strings.TrimSpace(input)

		switch input {
		case "/claude", "/restart", "/start":
			fmt.Printf("\n%s🚀 Opening Claude Code...%s\n\n", ColorGreen, ColorReset)
			return // Exits the shell, returning to the main loop to launch Claude

		case "/login":
			fmt.Printf("\n%s⚠️ Forcing new authentication...%s\n", ColorYellow, ColorReset)
			fmt.Printf("\n%s🔗 Fallback link: %s\n%s", ColorDark, ssoURL(DefaultPort), ColorReset)
			newToken := authenticateViaSSO()
			if newToken != "" {
				saveToken(newToken)
				fetchAndMergeSettings(newToken)
				ensureCliExists(newToken)
				fmt.Printf("\n%s✅ Connected! Launching Claude Code...%s\n\n", ColorGreen, ColorReset)
				return // Auto-launch Claude Code after successful login
			} else {
				fmt.Printf("\n%s❌ Authentication failed. Try again with /login.%s\n", ColorRed, ColorReset)
			}

		case "/reload", "/refresh":
			fmt.Printf("\n%s🔄 Reloading enterprise settings from server...%s\n", ColorYellow, ColorReset)
			currentToken := getToken()
			fetchAndMergeSettings(currentToken)
			fmt.Printf("\n%s✅ Settings up to date!%s\n", ColorGreen, ColorReset)

		case "/sync-models":
			fmt.Printf("\n%s📡 Syncing available AI models from Enterprise Proxy...%s\n", ColorYellow, ColorReset)
			currentToken := getToken()
			if isTokenValid(currentToken) {
				fetchAndMergeSettings(currentToken)
				fmt.Printf("\n%s✅ Models synced successfully.%s\n", ColorGreen, ColorReset)
			} else {
				fmt.Printf("\n%s❌ Cannot sync models. Token invalid. Please type /login.%s\n", ColorRed, ColorReset)
			}

		case "/usage":
			fmt.Printf("\n%s📊 Enterprise Usage Statistics:%s\n", ColorDark, ColorReset)
			fmt.Printf("   Please check your developer dashboard to see your token consumption:\n")
			fmt.Printf("   %s👉 %s/dashboard%s\n\n", ColorGreen, CliDownloadBase, ColorReset)

		case "/about":
			printLogo(isLoggedIn)
			fmt.Printf("\n%sWrapper AI Wrapper for Claude Code - Enterprise Edition%s\n", ColorDark, ColorReset)
			fmt.Printf("\n%sVersion 1.0.0 | Built for AI Engineering & DevOps%s\n\n", ColorDark, ColorReset)

		case "/help":
			fmt.Printf("\n%s🛠️  Available Commands:%s\n", ColorDark, ColorReset)
			fmt.Printf("  %s/claude%s       - Open Claude Code\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/restart%s      - Launch the Claude Code interface again\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/login%s        - Force a new SSO authentication session\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/reload%s       - Refresh enterprise settings (alias: /refresh)\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/sync-models%s  - Fetch latest available models from proxy\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/usage%s        - View usage statistics and token consumption\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/about%s        - Show tool version and information\n", ColorGreen, ColorReset)
			fmt.Printf("  %s/exit%s         - Close Wrapper completely\n\n", ColorGreen, ColorReset)

		case "/exit", "/quit":
			fmt.Printf("\n%s[Wrapper] Shutting down. Goodbye!%s\n\n\n", ColorYellow, ColorReset)
			os.Exit(0)

		case "":
			// User pressed Enter, do nothing

		default:
			fmt.Printf("\n%s❌ Unknown command '%s'. Type /help for the list of commands.%s\n", ColorRed, input, ColorReset)
		}
	}
}

// --- UI & Animations ---

// colorize wraps text in a true-color ANSI code
func colorize(text string, r, g, b int, bold bool) string {
	if bold {
		return fmt.Sprintf("\033[1;38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}

// getTermWidth returns terminal width via ioctl syscall
func getTermWidth() int {
	type winsize struct {
		Row, Col, Xpixel, Ypixel uint16
	}
	ws := &winsize{}
	ret, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(syscall.Stdout),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if int(ret) == -1 || ws.Col == 0 {
		return 80
	}
	w := int(ws.Col)
	if w < 50 {
		w = 50
	}
	return w
}

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(s string) string {
	r := s
	for strings.Contains(r, "\033[") {
		start := strings.Index(r, "\033[")
		end := strings.Index(r[start:], "m")
		if end == -1 {
			break
		}
		r = r[:start] + r[start+end+1:]
	}
	return r
}

func printLogo(isLoggedIn bool) {
	// Two clean colors only
	red := func(s string) string { return colorize(s, 240, 0, 11, false) }     // rgb(240,0,11)
	redBold := func(s string) string { return colorize(s, 240, 0, 11, true) }  // rgb(240,0,11) bold
	gray := func(s string) string { return colorize(s, 150, 150, 150, false) } // clean gray

	totalW := getTermWidth()
	if totalW > 100 {
		totalW = 100
	}

	rightW := 28
	leftW := totalW - rightW - 3
	if leftW < 30 {
		leftW = 30
		totalW = leftW + rightW + 3
	}

	v := "1.0.0"
	pipe := red("│")

	// Top border
	titleStr := " Wrapper AI v" + v + " "
	dashR := totalW - 5 - len([]rune(titleStr))
	if dashR < 1 {
		dashR = 1
	}

	fmt.Println()
	fmt.Printf("  %s%s%s\n",
		red("╭───"),
		redBold(titleStr),
		red(strings.Repeat("─", dashR)+"╮"))

	// Content
	type line struct {
		styled, plain string
	}

	mascot := []string{"", "▐▛███▜▌", "▝▜█████▛▘", "  ▘▘ ▝▝", ""}

	left := []line{
		{"\033[1mWelcome to Wrapper!\033[0m", "Welcome to Wrapper!"},
		{"", ""},
	}
	for _, m := range mascot {
		left = append(left, line{redBold(m), m})
	}
	left = append(left,
		line{gray("Claude Code · powered by Wrapper"), "Claude Code · powered by Wrapper"},
		line{gray("~/.claude"), "~/.claude"},
	)

	// Login status indicator
	var statusLine string
	if isLoggedIn {
		statusLine = colorize("● Connected", 76, 175, 80, true)
	} else {
		statusLine = colorize("○ Not connected", 240, 0, 11, true)
	}

	right := []string{
		redBold("Status"),
		statusLine,
		"",
		redBold("Tips for getting started"),
		gray("Type /help for commands"),
		gray("Type /about for info"),
	}

	for len(right) < len(left) {
		right = append(right, "")
	}
	for len(left) < len(right) {
		left = append(left, line{"", ""})
	}

	for i := 0; i < len(left); i++ {
		// Left: centered
		pLen := len([]rune(left[i].plain))
		var sl string
		if pLen == 0 {
			sl = strings.Repeat(" ", leftW)
		} else {
			pad := (leftW - pLen) / 2
			if pad < 0 {
				pad = 0
			}
			rp := leftW - pLen - pad
			if rp < 0 {
				rp = 0
			}
			sl = strings.Repeat(" ", pad) + left[i].styled + strings.Repeat(" ", rp)
		}

		// Right: left-aligned, padded
		rc := stripAnsi(right[i])
		rPad := rightW - len([]rune(rc))
		if rPad < 0 {
			rPad = 0
		}
		sr := right[i] + strings.Repeat(" ", rPad)

		fmt.Printf("  %s%s%s%s%s\n", pipe, sl, pipe, sr, pipe)
	}

	// Bottom border
	fmt.Printf("  %s\n", red("╰"+strings.Repeat("─", totalW-2)+"╯"))
	fmt.Println()
}

func showSpinner(text string, done chan bool) {
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r\033[K")
			return
		default:
			fmt.Printf("\r%s%s%s %s", ColorRed, chars[i], ColorReset, text)
			i = (i + 1) % len(chars)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// --- Claude Code Binary Management (Download & Execution) ---

func getCliPath() string {
	home, _ := os.UserHomeDir()
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return filepath.Join(home, ".claude", "bin", "claude"+ext)
}

func getCliDownloadURL() string {
	switch runtime.GOOS {
	case "windows":
		return CliDownloadBase + "/cli/win"
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return CliDownloadBase + "/cli/mac-m-series"
		}
		return CliDownloadBase + "/cli/mac-intel"
	case "linux":
		return CliDownloadBase + "/cli/linux"
	default:
		return ""
	}
}

func ensureCliExists(token string) {
	cliPath := getCliPath()

	if _, err := os.Stat(cliPath); err == nil {
		return // Binary already exists
	}

	downloadURL := getCliDownloadURL()
	if downloadURL == "" {
		fmt.Printf("\n%s❌ OS not supported for automatic download.%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	doneSpinner := make(chan bool)
	go showSpinner("Downloading Claude engine (first use)...", doneSpinner)

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		doneSpinner <- true
		fmt.Printf("\n%s❌ Request error: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)

	doneSpinner <- true

	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("\n%s❌ Download failed from %s (HTTP %d)%s\n", ColorRed, downloadURL, resp.StatusCode, ColorReset)
		os.Exit(1)
	}
	defer resp.Body.Close()

	os.MkdirAll(filepath.Dir(cliPath), 0755)
	out, err := os.Create(cliPath)
	if err != nil {
		fmt.Printf("\n%s❌ Failed to create local file: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("\n%s❌ Error writing to file: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	if runtime.GOOS != "windows" {
		os.Chmod(cliPath, 0755)
	}

	fmt.Printf("\n%s✅ Engine downloaded successfully!%s\n\n", ColorGreen, ColorReset)
}

func launchClaudeCode() {
	cliPath := getCliPath()
	args := os.Args[1:]

	cmd := exec.Command(cliPath, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// Program exits silently if Claude Code returns an exit code
	}
}

// --- Background Logic (Renewal Daemon) ---

func startBackgroundTokenCheck() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		token := getToken()
		if !isTokenValid(token) {
			fmt.Printf("\n\n%s[Wrapper] ⚠️ Token expired. Launching SSO login...%s\n", ColorYellow, ColorReset)
			fmt.Printf("\n%s[Wrapper] 🔗 Manual reconnection link: %s\n%s", ColorDark, ssoURL(DefaultPort), ColorReset)

			newToken := authenticateViaSSO()
			if newToken != "" {
				saveToken(newToken)
				fetchAndMergeSettings(newToken)
				fmt.Printf("\n%s[Wrapper] ✅ Token renewed! (Note: re-run your last command if it failed)%s\n> ", ColorGreen, ColorReset)
			}
		}
	}
}

func isTokenValid(token string) bool {
	if token == "" {
		return false
	}
	req, err := http.NewRequest("GET", ProxyBaseURL+"/models", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode == http.StatusUnauthorized {
		return false
	}
	return true
}

// --- SSO Management ---

func authenticateViaSSO() string {
	for {
		token := attemptSSO()
		if token != "" {
			return token
		}

		// No token received — ask if the user wants to retry
		// add print barre separe
		fmt.Printf("\n%s──────────────────────────────────────────────────────────────────────────────────────────────────%s\n", ColorDark, ColorReset)
		fmt.Printf("\n%s❌ Authentication failed — no token received.%s\n", ColorRed, ColorReset)
		fmt.Printf("\n%s⚠️  You are not connected. Would you like to retry? (y/n): %s", ColorYellow, ColorReset)

		reader := bufio.NewReader(os.Stdin)
		var answerBytes []byte
		for {
			b, err := reader.ReadByte()
			if err != nil || b == '\n' || b == '\r' {
				break
			}
			answerBytes = append(answerBytes, b)
		}
		answer := strings.TrimSpace(strings.ToLower(string(answerBytes)))

		if answer != "y" && answer != "yes" {
			return ""
		}

		fmt.Printf("\n%s🔄 Retrying authentication...%s\n", ColorGreen, ColorReset)
	}
}

func attemptSSO() string {
	tokenChan := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token != "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body style="font-family:sans-serif;text-align:center;padding:50px;background:#1a1a2e;color:#e0e0e0;">
				<h1 style="color:#4caf50;">Wrapper: Authentication successful!</h1>
				<p>You can close this tab and return to your terminal.</p>
			</body></html>`)
			tokenChan <- token
		} else {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body style="font-family:sans-serif;text-align:center;padding:50px;background:#1a1a2e;color:#e0e0e0;">
				<h1 style="color:#f44336;">Error: No token received</h1>
				<p>The SSO server did not return a valid token.</p>
				<p>Return to your terminal and type <b>/login</b> to try again.</p>
			</body></html>`)
			fmt.Printf("\n%sSSO callback received but no token was provided by the server.%s\n", ColorRed, ColorReset)
			tokenChan <- ""
		}
	})

	// Try ports from DefaultPort to DefaultPort+MaxPortRetries
	var server *http.Server
	port := 0
	for p := DefaultPort; p < DefaultPort+MaxPortRetries; p++ {
		addr := fmt.Sprintf("127.0.0.1:%d", p)
		server = &http.Server{Addr: addr, Handler: mux}

		ln, err := net.Listen("tcp", addr)
		if err != nil {
			// fmt.Printf("\n%s⚠️ Port %d busy, trying %d...%s\n", ColorYellow, p, p+1, ColorReset)
			continue
		}

		port = p
		go func() {
			if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
				tokenChan <- ""
			}
		}()
		break
	}

	if port == 0 {
		fmt.Printf("\n%s❌ Error: All ports %d-%d are in use.%s\n", ColorRed, DefaultPort, DefaultPort+MaxPortRetries-1, ColorReset)
		return ""
	}

	url := ssoURL(port)
	fmt.Printf("\n%sIf the browser doesn't open, please click here:%s\n", ColorDark, ColorReset)
	fmt.Printf("%s🔗 %s%s\n", ColorRed, url, ColorReset)
	fmt.Printf("\n%sWaiting for login in the browser...%s\n", ColorYellow, ColorReset)
	openBrowser(url)

	select {
	case token := <-tokenChan:
		server.Shutdown(context.Background())
		return token
	case <-time.After(5 * time.Minute):
		fmt.Printf("\n%s⏰ Timeout exceeded (5 minutes).%s\n", ColorRed, ColorReset)
		server.Shutdown(context.Background())
		return ""
	}
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	_ = err
}

// --- JSON Management (Claude Code Settings) ---

func getSettingsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "settings.json")
}

func readSettings() map[string]interface{} {
	path := getSettingsPath()
	settings := make(map[string]interface{})
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &settings)
	}
	return settings
}

func getToken() string {
	settings := readSettings()
	if envMap, ok := settings["env"].(map[string]interface{}); ok {
		if token, ok := envMap["ANTHROPIC_API_KEY"].(string); ok {
			return token
		}
	}
	return ""
}

func saveToken(token string) {
	path := getSettingsPath()
	os.MkdirAll(filepath.Dir(path), 0755)

	settings := readSettings()

	envMap, ok := settings["env"].(map[string]interface{})
	if !ok {
		envMap = make(map[string]interface{})
	}

	envMap["ANTHROPIC_API_KEY"] = token
	envMap["ANTHROPIC_BASE_URL"] = ProxyBaseURL

	settings["env"] = envMap

	data, _ := json.MarshalIndent(settings, "", "  ")
	os.WriteFile(path, data, 0644)
}

func fetchAndMergeSettings(token string) {
	req, err := http.NewRequest("GET", ConfigAPIURL, nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	apiSettings := make(map[string]interface{})
	if err := json.NewDecoder(resp.Body).Decode(&apiSettings); err != nil {
		return
	}

	path := getSettingsPath()
	currentSettings := readSettings()

	for key, value := range apiSettings {
		if key == "env" {
			if incomingEnv, ok := value.(map[string]interface{}); ok {
				currentEnv, exists := currentSettings["env"].(map[string]interface{})
				if !exists {
					currentEnv = make(map[string]interface{})
				}
				for k, v := range incomingEnv {
					currentEnv[k] = v
				}
				currentSettings["env"] = currentEnv
			}
		} else {
			currentSettings[key] = value
		}
	}

	data, _ := json.MarshalIndent(currentSettings, "", "  ")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, data, 0644)
}
