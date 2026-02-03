package interactive

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/eiannone/keyboard"
	"github.com/pefman/sidekick/internal/config"
	"github.com/pefman/sidekick/internal/ollama"
	"github.com/pefman/sidekick/internal/updater"
)

// formatSize converts bytes to human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type InteractiveMode struct {
	config *config.Config
	reader *bufio.Reader
	mu     sync.RWMutex

	updateChecked   bool
	updateAvailable bool
	updateVersion   string
}

func New() *InteractiveMode {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.GetDefault()
	}

	return &InteractiveMode{
		config: cfg,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (im *InteractiveMode) Run() error {
	im.checkForUpdatesAsync()

	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()

	modes := []string{"ask", "edit", "plan"}
	modeIdx := 0
	selectedIdx := -1
	var input []rune

	for {
		items := []MenuItem{
			{Label: "Settings", Value: "settings"},
			{Label: "Models", Value: "models"},
			{Label: "Help", Value: "help"},
			{Label: "Quit", Value: "quit"},
		}

		if im.hasUpdate() {
			label := "Update Available"
			if version := im.getUpdateVersion(); version != "" {
				label = fmt.Sprintf("Update Available (%s)", version)
			}
			items = append(items[:2], append([]MenuItem{{Label: label, Value: "update"}}, items[2:]...)...)
		}

		if selectedIdx >= len(items) {
			selectedIdx = len(items) - 1
		}

		im.clearScreen()
		im.showWelcome()
		if selectedIdx == -1 {
			fmt.Printf("%s  prompt >%s %s\n\n", orange, reset, string(input))
		} else {
			fmt.Printf("  prompt > %s\n\n", string(input))
		}
		for i, item := range items {
			if i == selectedIdx {
				fmt.Printf("%s▸ %s%s\n", orange, item.Label, reset)
			} else {
				fmt.Printf("  %s\n", item.Label)
			}
		}

		fmt.Println()
		fmt.Printf("  Mode: %s%s%s  (Tab to change, Enter to submit, Esc to quit)\n", orange, strings.ToUpper(modes[modeIdx]), reset)
		fmt.Println("  Menu: Use ↑↓ to select, Enter to open/execute")
		fmt.Println()

		char, key, err := keyboard.GetKey()
		if err != nil {
			return err
		}

		switch key {
		case keyboard.KeyEsc, keyboard.KeyArrowLeft:
			im.clearScreen()
			fmt.Printf("\n%s▸%s Goodbye!\n\n", orange, reset)
			return nil
		case keyboard.KeyArrowUp:
			if selectedIdx > -1 {
				selectedIdx--
			}
		case keyboard.KeyArrowDown:
			if selectedIdx < len(items)-1 {
				selectedIdx++
			}
		case keyboard.KeyTab:
			modeIdx = (modeIdx + 1) % len(modes)
		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case keyboard.KeySpace:
			input = append(input, ' ')
		case keyboard.KeyEnter:
			if selectedIdx == -1 {
				prompt := strings.TrimSpace(string(input))
				if prompt == "" {
					break
				}
				keyboard.Close()
				customPrompt := fmt.Sprintf("MODE: %s\n%s", strings.ToUpper(modes[modeIdx]), prompt)
				if err := im.runPrompt(customPrompt); err != nil {
					fmt.Printf("\n%s✗%s Error: %v\n", orange, reset, err)
					im.pressEnterToContinue()
				}
				input = []rune{}
				if err := keyboard.Open(); err != nil {
					return err
				}
				break
			}

			switch items[selectedIdx].Value {
			case "settings":
				keyboard.Close()
				im.settingsMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "models":
				keyboard.Close()
				im.modelsMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "update":
				keyboard.Close()
				im.updateMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "help":
				keyboard.Close()
				im.showHelp()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "quit":
				im.clearScreen()
				fmt.Printf("\n%s▸%s Goodbye!\n\n", orange, reset)
				return nil
			}
		case keyboard.KeyArrowRight:
			switch items[selectedIdx].Value {
			case "settings":
				keyboard.Close()
				im.settingsMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "models":
				keyboard.Close()
				im.modelsMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "update":
				keyboard.Close()
				im.updateMenu()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "help":
				keyboard.Close()
				im.showHelp()
				if err := keyboard.Open(); err != nil {
					return err
				}
			case "quit":
				im.clearScreen()
				fmt.Printf("\n%s▸%s Goodbye!\n\n", orange, reset)
				return nil
			}
		default:
			if key == 0 && char != 0 {
				input = append(input, char)
			}
		}
	}
}

func (im *InteractiveMode) runPrompt(customPrompt string) error {
	im.clearScreen()
	im.showWelcome()

	// Get scan path
	fmt.Printf("\n%s▸%s Path (press Enter for current directory): ", orange, reset)
	path := im.readInput()
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Use config settings
	model := im.config.DefaultModel
	scanType := "custom"

	// Start scan immediately
	fmt.Println()
	if err := performScan(path, model, im.config.Debug, scanType, customPrompt); err != nil {
		return err
	}

	im.pressEnterToContinue()
	return nil
}

func (im *InteractiveMode) clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func (im *InteractiveMode) checkForUpdatesAsync() {
	latest, hasUpdate, err := updater.CheckForUpdate()
	if err != nil {
		im.mu.Lock()
		im.updateChecked = true
		im.mu.Unlock()
		return
	}

	im.mu.Lock()
	im.updateChecked = true
	im.updateAvailable = hasUpdate
	if hasUpdate && latest != nil {
		im.updateVersion = latest.Version()
	}
	im.mu.Unlock()
}

func (im *InteractiveMode) hasUpdate() bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.updateChecked && im.updateAvailable
}

func (im *InteractiveMode) getUpdateVersion() string {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.updateVersion
}

const (
	orange = "\033[38;5;208m"
	cyan   = "\033[38;5;51m"
	gray   = "\033[38;5;240m"
	reset  = "\033[0m"
	bold   = "\033[1m"
)

func (im *InteractiveMode) showWelcome() {
	fmt.Printf("%s%s▸ SIDEKICK%s %sAI Code Assistant%s\n\n", bold, orange, reset, gray, reset)
}

func (im *InteractiveMode) showMainMenu() {
	fmt.Printf("%s[P]%s Prompt\n", orange, reset)
	fmt.Printf("%s[T]%s Settings  %s[M]%s Models  %s[H]%s Help  %s[Q]%s Quit\n",
		orange, reset, orange, reset, orange, reset, orange, reset)
	fmt.Printf("%s▸%s ", orange, reset)
}

func (im *InteractiveMode) readInput() string {
	input, _ := im.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (im *InteractiveMode) readPromptWithMode() (string, string, bool) {
	modes := []string{"ask", "edit", "plan"}
	modeIdx := 0
	var input []rune
	skipInitialEnter := true

	if err := keyboard.Open(); err != nil {
		return "", "", false
	}
	defer keyboard.Close()

	for {
		im.clearScreen()
		im.showWelcome()
		fmt.Printf("%s▸ PROMPT%s\n\n", orange, reset)
		fmt.Printf("%sMode:%s %s%s%s  (Tab to change, Enter to submit, Esc/← to cancel)\n\n",
			gray, reset, orange, strings.ToUpper(modes[modeIdx]), reset)
		fmt.Printf("%s▸%s %s", orange, reset, string(input))

		char, key, err := keyboard.GetKey()
		if err != nil {
			return "", "", false
		}

		switch key {
		case keyboard.KeyEsc, keyboard.KeyArrowLeft:
			return "", "", false
		case keyboard.KeyEnter:
			if skipInitialEnter && len(input) == 0 {
				skipInitialEnter = false
				continue
			}
			prompt := strings.TrimSpace(string(input))
			if prompt == "" {
				continue
			}
			return prompt, modes[modeIdx], true
		case keyboard.KeyBackspace, keyboard.KeyBackspace2:
			skipInitialEnter = false
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case keyboard.KeyTab:
			skipInitialEnter = false
			modeIdx = (modeIdx + 1) % len(modes)
		default:
			skipInitialEnter = false
			if key == 0 && char != 0 {
				input = append(input, char)
			}
		}
	}
}

func (im *InteractiveMode) pressEnterToContinue() {
	fmt.Print("\nPress Enter to continue...")
	im.reader.ReadString('\n')
	im.clearScreen()
	im.showWelcome()
}

func (im *InteractiveMode) waitForBack() {
	if err := keyboard.Open(); err != nil {
		im.pressEnterToContinue()
		return
	}
	defer keyboard.Close()

	for {
		_, key, err := keyboard.GetKey()
		if err != nil {
			break
		}
		switch key {
		case keyboard.KeyEnter, keyboard.KeyEsc, keyboard.KeyArrowLeft:
			return
		}
	}
}

func (im *InteractiveMode) scanMenu() error {
	// Prompt input
	promptText, mode, ok := im.readPromptWithMode()
	if !ok {
		return nil
	}
	customPrompt := fmt.Sprintf("MODE: %s\n%s", strings.ToUpper(mode), promptText)

	im.clearScreen()
	im.showWelcome()
	fmt.Printf("%s▸ PROMPT%s\n", orange, reset)

	// Get scan path
	fmt.Printf("\n%s▸%s Path (press Enter for current directory): ", orange, reset)
	path := im.readInput()
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// Use config settings
	model := im.config.DefaultModel
	scanType := "custom"

	// Start scan immediately
	fmt.Println()
	if err := performScan(path, model, im.config.Debug, scanType, customPrompt); err != nil {
		return err
	}

	im.pressEnterToContinue()
	return nil
}

func (im *InteractiveMode) settingsMenu() {
	for {
		items := []MenuItem{
			{Label: fmt.Sprintf("URL: %s", im.config.OllamaURL), Value: "url"},
			{Label: fmt.Sprintf("Debug: %v", im.config.Debug), Value: "debug"},
		}

		selected, err := SelectMenu("SETTINGS", items, 0)
		if err != nil || selected == -1 {
			return
		}

		switch items[selected].Value {
		case "debug":
			im.config.Debug = !im.config.Debug
			im.config.Save()
		case "reset":
			im.config = config.GetDefault()
			im.config.Save()
			im.clearScreen()
			im.showWelcome()
			fmt.Printf("\n%s✓%s Settings reset to defaults\n", orange, reset)
			im.pressEnterToContinue()
		case "back":
			return
		}
	}
}

func (im *InteractiveMode) changeOllamaURL() bool {
	fmt.Printf("\n🔗 Current URL: %s\n", im.config.OllamaURL)
	fmt.Print("Enter new Ollama URL: ")
	url := im.readInput()
	if url != "" {
		// Validate URL format
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			fmt.Println("\n❌ Invalid URL: must start with http:// or https://")
			im.pressEnterToContinue()
			return false
		}
		// Warn if using HTTP for non-localhost
		if strings.HasPrefix(url, "http://") && !strings.Contains(url, "localhost") && !strings.Contains(url, "127.0.0.1") {
			fmt.Printf("\n%s⚠️  Warning: Using HTTP for non-localhost connection is insecure%s\n", orange, reset)
			fmt.Print("Continue anyway? (y/N): ")
			confirm := im.readInput()
			if confirm != "y" && confirm != "Y" {
				return false
			}
		}
		im.config.OllamaURL = url
		if err := im.config.Save(); err != nil {
			fmt.Printf("\n❌ Failed to save: %v\n", err)
			im.pressEnterToContinue()
			return false
		}
		fmt.Println("\n✅ Ollama URL updated - discovering models...")
		return true
	}
	return false
}

func (im *InteractiveMode) modelsMenu() {
	statusMessage := ""
	for {
		client := ollama.NewClient(im.config.OllamaURL)
		models, err := client.ListModelsWithDetails()
		if err != nil {
			fmt.Printf("\n❌ Error: %v\n", err)
			im.pressEnterToContinue()
			return
		}

		if len(models) == 0 {
			im.clearScreen()
			im.showWelcome()
			fmt.Println("\n⚠️  No models found. Install a model first:")
			fmt.Println("   ollama pull qwen2.5-coder:14b")
			fmt.Println("   ollama pull deepseek-r1:14b")
			im.pressEnterToContinue()
			return
		}

		// Sort models alphabetically (case-insensitive)
		sort.Slice(models, func(i, j int) bool {
			return strings.ToLower(models[i].Name) < strings.ToLower(models[j].Name)
		})

		// Build menu items
		items := make([]MenuItem, len(models)+2)
		currentIdx := 0
		for i, model := range models {
			sizeStr := formatSize(model.Size)
			prefix := "  "
			if model.Name == im.config.DefaultModel {
				prefix = "✓ "
				currentIdx = i
			}
			items[i] = MenuItem{
				Label: fmt.Sprintf("%s%-40s     %s%s%s", prefix, model.Name, gray, sizeStr, reset),
				Value: model.Name,
			}
		}
		// Add URL change and back options
		items[len(models)] = MenuItem{
			Label: fmt.Sprintf("\n🔗 Change Ollama URL (Current: %s)", im.config.OllamaURL),
			Value: "__change_url__",
		}
		items[len(models)+1] = MenuItem{
			Label: "← Back",
			Value: "__back__",
		}

		// Add helpful note at the top
		im.clearScreen()
		im.showWelcome()
		fmt.Printf("%s▸ MODELS%s\n\n", orange, reset)
		if statusMessage != "" {
			fmt.Printf("%s%s%s\n\n", orange, statusMessage, reset)
			statusMessage = ""
		}
		fmt.Printf("%sRecommended for security scanning:%s\n", gray, reset)
		fmt.Printf("  %s●%s qwen2.5-coder:32b - Best accuracy & JSON compliance\n", cyan, reset)
		fmt.Printf("  %s●%s deepseek-coder-v2:16b - Fast & consistent\n", cyan, reset)
		fmt.Printf("  %s●%s deepseek-r1:14b+ - Most thorough analysis\n", cyan, reset)
		fmt.Printf("  %s●%s qwen2.5-coder:7b - Speed (large codebases)\n\n", cyan, reset)

		selected, err := SelectMenu("Select default model", items, currentIdx)
		if err != nil || selected == -1 {
			return
		}

		// Handle special menu items
		if items[selected].Value == "__back__" {
			return
		}
		if items[selected].Value == "__change_url__" {
			if im.changeOllamaURL() {
				// URL changed successfully, refresh models
				continue
			}
			// URL change failed or cancelled, stay in menu
			continue
		}

		// Set as default
		modelName := items[selected].Value
		im.config.DefaultModel = modelName
		if err := im.config.Save(); err != nil {
			fmt.Printf("\n❌ Failed to save: %v\n", err)
			im.pressEnterToContinue()
		} else {
			statusMessage = fmt.Sprintf("✓ Default model set to: %s", modelName)
			continue
		}
	}
}

func (im *InteractiveMode) updateMenu() {
	im.clearScreen()
	im.showWelcome()
	fmt.Printf("%s▸ UPDATE%s\n\n", orange, reset)

	fmt.Printf("Current version: %s%s%s\n\n", cyan, updater.Version, reset)
	fmt.Println("🔍 Checking for updates...")

	latest, hasUpdate, err := updater.CheckForUpdate()
	if err != nil {
		fmt.Printf("%s✗%s Error checking for updates: %v\n", orange, reset, err)
		im.pressEnterToContinue()
		return
	}

	if !hasUpdate {
		fmt.Printf("\n%s✓%s You're already running the latest version!\n", cyan, reset)
		im.pressEnterToContinue()
		return
	}

	fmt.Printf("\n%s📦 New version available: %s%s\n", orange, latest.Version(), reset)
	fmt.Println("\nUpdating now...")
	if err := updater.Update(); err != nil {
		fmt.Printf("\n%s✗%s Update failed: %v\n", orange, reset, err)
		fmt.Println("\nYou can also try updating via CLI:")
		fmt.Println("  sudo sidekick update")
	} else {
		fmt.Printf("\n%s✓%s Please restart sidekick to use the new version.\n", cyan, reset)
	}

	im.pressEnterToContinue()
}

func (im *InteractiveMode) showHelp() {
	im.clearScreen()
	im.showWelcome()
	fmt.Printf("%s▸ HELP%s\n", orange, reset)
	fmt.Println()
	fmt.Println("FEATURES:")
	fmt.Println("  • Prompt the AI about your code")
	fmt.Println("  • Uses local LLM via Ollama (privacy-first)")
	fmt.Println("  • Supports multiple programming languages")
	fmt.Println("  • Customizable settings and models")
	fmt.Println("  • Auto-update support from GitHub releases")
	fmt.Println()
	fmt.Println("COMMAND LINE MODE:")
	fmt.Println("  sidekick scan [path]           # Scan a directory")
	fmt.Println("  sidekick scan -m model         # Use specific model")
	fmt.Println("  sidekick update                # Run update check")
	fmt.Println("  sidekick install               # Install to system")
	fmt.Println("  sidekick --version             # Show version")
	fmt.Println()
	fmt.Println("INTERACTIVE MODE:")
	fmt.Println("  sidekick                       # Launch interactive UI")
	fmt.Println()
	fmt.Println("KEYBOARD SHORTCUTS:")
	fmt.Println("  Use ↑↓ arrows to navigate")
	fmt.Println("  Enter/→ to select")
	fmt.Println("  ←/Esc to go back")
	fmt.Println()
	fmt.Printf("Version: %s%s%s\n", cyan, updater.Version, reset)
	fmt.Println("GitHub: https://github.com/pefman/sidekick")

	fmt.Print("\nPress Enter or ←/Esc to go back...")
	im.waitForBack()
	im.clearScreen()
	im.showWelcome()
}
