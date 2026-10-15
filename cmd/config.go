package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/veenoise/whaeh/pkg/config"
	"golang.org/x/term"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage whaeh configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return cmd.Help()
		}
		return runInteractiveConfig(cmd)
	},
}

var errConfigCancelled = errors.New("configuration cancelled")

func runInteractiveConfig(cmd *cobra.Command) error {
	doc, err := config.LoadFile()
	if err != nil {
		return err
	}

	if len(doc.Profiles) == 1 {
		doc.ActiveProfile = config.ProfileNames(doc)[0]
	}
	if len(doc.Profiles) == 1 {
		if _, err := config.SaveFile(doc); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
	}

	for {
		choice, err := runProfileMenu(cmd.InOrStdin(), cmd.OutOrStdout(), doc)
		if err != nil {
			return err
		}
		if choice.exit {
			return nil
		}
		if choice.profile == -1 {
			if err := createProfile(&doc, cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
				if errors.Is(err, errConfigCancelled) {
					return nil
				}
				return err
			}
			if _, err := config.SaveFile(doc); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			continue
		}

		names := config.ProfileNames(doc)
		name := names[choice.profile]
		switch choice.action {
		case 0:
			doc.ActiveProfile = name
		case 1:
			if err := editProfile(&doc, name, cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
				if errors.Is(err, errConfigCancelled) {
					return nil
				}
				return err
			}
		case 2:
			deleted, err := deleteProfile(&doc, name, cmd.InOrStdin(), cmd.OutOrStdout())
			if err != nil {
				if errors.Is(err, errConfigCancelled) {
					return nil
				}
				return err
			}
			if !deleted {
				continue
			}
			if len(doc.Profiles) > 0 && doc.ActiveProfile == "" {
				doc.ActiveProfile = config.ProfileNames(doc)[0]
			}
		}
		target, err := config.SaveFile(doc)
		if err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "\nConfiguration saved to %s\n", linkPath(target))
	}
}

type profileMenuChoice struct {
	profile int
	action  int
	exit    bool
}

type profileMenuModel struct {
	options []string
	cursor  int
	action  int
	chosen  profileMenuChoice
}

func (m profileMenuModel) Init() tea.Cmd { return nil }

func (m profileMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.options)-1 {
			m.cursor++
		}
	case "left", "h":
		if m.cursor > 0 && m.cursor < len(m.options)-1 && m.action > 0 {
			m.action--
		}
	case "right", "l":
		if m.cursor > 0 && m.cursor < len(m.options)-1 && m.action < 2 {
			m.action++
		}
	case "enter":
		if m.cursor == len(m.options)-1 {
			m.chosen.exit = true
		} else {
			m.chosen = profileMenuChoice{profile: m.cursor - 1, action: m.action}
		}
		return m, tea.Quit
	case "esc", "ctrl+c":
		m.chosen.exit = true
		return m, tea.Quit
	}
	return m, nil
}

func (m profileMenuModel) View() string {
	var b strings.Builder
	b.WriteString(clearScreen())
	fmt.Fprintf(&b, "\n%s\n", highlight("Select profile", colorCyan))
	for i, option := range m.options {
		marker := "  "
		if i == m.cursor {
			marker = highlight("> ", colorCyan)
			option = highlight(option, colorCyan)
		}
		fmt.Fprintf(&b, "%s%s\n", marker, option)
	}
	if m.cursor > 0 && m.cursor < len(m.options)-1 {
		b.WriteString("\n")
		for i, action := range []string{"Set Active", "Edit", "Delete"} {
			if i > 0 {
				b.WriteString("   ")
			}
			if i == m.action {
				fmt.Fprintf(&b, "[ %s ]", highlight(action, colorYellow))
			} else {
				fmt.Fprintf(&b, "  %s  ", action)
			}
		}
		b.WriteString("\nUse ↑/↓ for profiles, ←/→ for actions, Enter to select, Esc to exit.")
	} else {
		b.WriteString("\nUse ↑/↓ and Enter. Esc to exit.")
	}
	return b.String()
}

func runProfileMenu(in io.Reader, out io.Writer, doc config.File) (profileMenuChoice, error) {
	model := profileMenuModel{
		options: profileOptions(doc),
		chosen:  profileMenuChoice{profile: -2, action: -1},
	}
	result, err := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return profileMenuChoice{}, err
	}
	choice := result.(profileMenuModel).chosen
	if choice.profile == -2 && !choice.exit {
		return profileMenuChoice{}, fmt.Errorf("configuration cancelled")
	}
	return choice, nil
}

type selectorModel struct {
	title      string
	options    []string
	cursor     int
	chosen     int
	horizontal bool
}

func (m selectorModel) Init() tea.Cmd { return nil }

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if !m.horizontal && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if !m.horizontal && m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "left", "h":
			if m.horizontal && m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.horizontal && m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = m.cursor
			return m, tea.Quit
		case "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m selectorModel) View() string {
	var b strings.Builder
	b.WriteString(clearScreen())
	fmt.Fprintf(&b, "\n%s\n", highlight(m.title, colorCyan))
	if m.horizontal {
		for i, option := range m.options {
			if i > 0 {
				b.WriteString("   ")
			}
			if i == m.cursor {
				fmt.Fprintf(&b, "[ %s ]", highlight(option, colorYellow))
			} else {
				fmt.Fprintf(&b, "  %s  ", option)
			}
		}
		b.WriteString("\n\nUse ←/→ and Enter.")
		return b.String()
	}
	for i, option := range m.options {
		marker := "  "
		if i == m.cursor {
			marker = highlight("> ", colorCyan)
			option = highlight(option, colorCyan)
		}
		fmt.Fprintf(&b, "%s%s\n", marker, option)
	}
	b.WriteString("\nUse ↑/↓ and Enter. Esc to exit.")
	return b.String()
}

func runSelector(in io.Reader, out io.Writer, title string, options []string, cursor int) (int, error) {
	if len(options) == 0 {
		return 0, fmt.Errorf("no options available")
	}
	model := selectorModel{title: title, options: options, cursor: cursor, chosen: -1, horizontal: strings.HasPrefix(title, "Select action")}
	result, err := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return 0, err
	}
	selected := result.(selectorModel).chosen
	if selected < 0 {
		return 0, errConfigCancelled
	}
	return selected, nil
}

func profileOptions(doc config.File) []string {
	options := []string{"+ Create new profile"}
	for _, name := range config.ProfileNames(doc) {
		label := name
		if name == doc.ActiveProfile {
			label = highlight(label+" (active)", colorGreen)
		}
		options = append(options, label)
	}
	options = append(options, "Exit")
	return options
}

func profileIndex(doc config.File) int {
	for i, name := range config.ProfileNames(doc) {
		if name == doc.ActiveProfile {
			return i + 1
		}
	}
	return 0
}

const (
	colorReset  = "\x1b[0m"
	colorCyan   = "\x1b[36m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
)

func highlight(value, color string) string {
	if os.Getenv("TERM") == "dumb" {
		return value
	}
	return color + value + colorReset
}

func clearScreen() string {
	if os.Getenv("TERM") == "dumb" {
		return ""
	}
	return "\x1b[2J\x1b[H"
}

type inputModel struct {
	label     string
	defaultV  string
	value     []rune
	secret    bool
	cancelled bool
}

func (m inputModel) Init() tea.Cmd { return nil }

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "enter":
		return m, tea.Quit
	case "esc", "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "backspace", "ctrl+h":
		if len(m.value) > 0 {
			m.value = m.value[:len(m.value)-1]
		}
	default:
		m.value = append(m.value, key.Runes...)
	}
	return m, nil
}

func (m inputModel) View() string {
	display := string(m.value)
	if m.secret {
		display = strings.Repeat("•", len(m.value))
	}
	if display == "" && m.defaultV != "" {
		display = highlight("["+m.defaultV+"]", colorGreen)
	}
	return fmt.Sprintf("%s%s: %s", clearScreen(), highlight(m.label, colorCyan), display)
}

func prompt(in io.Reader, out io.Writer, label, defaultValue string) (string, error) {
	return promptInput(in, out, label, defaultValue, false)
}

func promptSecret(in io.Reader, out io.Writer, label, current string) (string, error) {
	value, err := promptInput(in, out, label+" (hidden; Enter keeps current)", "", true)
	if err != nil {
		return "", err
	}
	if value == "" {
		return current, nil
	}
	return value, nil
}

func promptInput(in io.Reader, out io.Writer, label, defaultValue string, secret bool) (string, error) {
	model := inputModel{label: label, defaultV: defaultValue, secret: secret}
	result, err := tea.NewProgram(model, tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return "", err
	}
	input := result.(inputModel)
	fmt.Fprintln(out)
	if input.cancelled {
		return "", errConfigCancelled
	}
	if len(input.value) == 0 {
		return defaultValue, nil
	}
	return strings.TrimSpace(string(input.value)), nil
}

func createProfile(doc *config.File, in io.Reader, out io.Writer) error {
	name, err := prompt(in, out, "Profile name", "")
	if err != nil {
		return err
	}
	if name == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if _, exists := doc.Profiles[name]; exists {
		return fmt.Errorf("profile %q already exists", name)
	}
	profile, err := promptProfile(in, out, config.Profile{})
	if err != nil {
		return err
	}
	doc.Profiles[name] = profile
	if len(doc.Profiles) == 1 || doc.ActiveProfile == "" {
		doc.ActiveProfile = name
	}
	return nil
}

func promptProfile(in io.Reader, out io.Writer, current config.Profile) (config.Profile, error) {
	var err error
	if current.Provider, err = promptProvider(in, out, current.Provider); err != nil {
		return config.Profile{}, err
	}
	if current.APIKey, err = promptSecret(in, out, "API key", current.APIKey); err != nil {
		return config.Profile{}, err
	}
	if current.Model, err = prompt(in, out, "Model", current.Model); err != nil {
		return config.Profile{}, err
	}
	if current.BaseURL, err = prompt(in, out, "Base URL (optional)", current.BaseURL); err != nil {
		return config.Profile{}, err
	}
	if current.SystemPrompt == "" {
		current.SystemPrompt = config.DefaultSystemPrompt
	}
	if current.SystemPrompt, err = prompt(in, out, "System prompt", current.SystemPrompt); err != nil {
		return config.Profile{}, err
	}
	return current, nil
}

func promptProvider(in io.Reader, out io.Writer, current string) (string, error) {
	options := []string{"ollama", "openai", "gemini"}
	cursor := 0
	for i, option := range options {
		if option == current {
			cursor = i
			break
		}
	}
	if current != "" && options[cursor] != current {
		options = append([]string{current}, options...)
		cursor = 0
	}
	selected, err := runSelector(in, out, "Select provider", options, cursor)
	if err != nil {
		return "", err
	}
	return options[selected], nil
}

func editProfile(doc *config.File, oldName string, in io.Reader, out io.Writer) error {
	newName, err := prompt(in, out, "Profile name", oldName)
	if err != nil {
		return err
	}
	if newName == "" {
		return fmt.Errorf("profile name cannot be empty")
	}
	if newName != oldName {
		if _, exists := doc.Profiles[newName]; exists {
			return fmt.Errorf("profile %q already exists", newName)
		}
	}
	profile, err := promptProfile(in, out, doc.Profiles[oldName])
	if err != nil {
		return err
	}
	delete(doc.Profiles, oldName)
	doc.Profiles[newName] = profile
	if doc.ActiveProfile == oldName {
		doc.ActiveProfile = newName
	}
	return nil
}

func deleteProfile(doc *config.File, name string, in io.Reader, out io.Writer) (bool, error) {
	answer, err := prompt(in, out, fmt.Sprintf("Are you sure you want to delete %s? [y/N]", name), "")
	if err != nil {
		return false, err
	}
	if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" {
		return false, nil
	}
	wasActive := doc.ActiveProfile == name
	delete(doc.Profiles, name)
	if wasActive {
		doc.ActiveProfile = ""
		if len(doc.Profiles) > 0 {
			doc.ActiveProfile = config.ProfileNames(*doc)[0]
		}
	}
	return true, nil
}

func editInEditor(currentPrompt string, reader *bufio.Reader, cmd *cobra.Command) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("vi"); err == nil {
			editor = "vi"
		} else {
			return currentPrompt, fmt.Errorf("no editor found; set $EDITOR")
		}
	}

	tmpFile, err := os.CreateTemp("", "whaeh-system-prompt-*.txt")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(currentPrompt); err != nil {
		_ = tmpFile.Close()
		return "", err
	}
	_ = tmpFile.Close()

	c := exec.Command(editor, tmpPath)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(data))
	if result == "" {
		fmt.Fprint(cmd.OutOrStdout(), "File is empty. Reset to default prompt or leave empty? [D/e]: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))
		if choice == "e" || choice == "empty" {
			return "", nil
		}
		return config.DefaultSystemPrompt, nil
	}

	return result, nil
}

func linkPath(path string) string {
	if os.Getenv("TERM") == "dumb" {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	host, _ := os.Hostname()
	return fmt.Sprintf("\x1b]8;;file://%s%s\x1b\\%s\x1b]8;;\x1b\\", host, abs, path)
}

func init() {
	origHelp := configCmd.HelpFunc()
	configCmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		path, _ := config.ConfigPath()
		c.Long = fmt.Sprintf(`Manage whaeh configuration.

Use ↑/↓ to select a profile. Selected profiles show Set Active, Edit, and
Delete actions. Use ←/→ to select an action, then press Enter. Select Exit or
press Esc to leave without saving an unfinished create or edit.

Configuration file:
  %s`, linkPath(path))
		origHelp(c, args)
	})
}
