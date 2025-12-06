package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	vaultDir string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory", err)
	}
	vaultDir = fmt.Sprintf("%s/.totion", homeDir)
}

type model struct {
	newFileInput           textinput.Model
	createFileInputVisible bool
	currentFile            *os.File
	noteTextArea           textarea.Model
}

func initialModel() model {
	err := os.MkdirAll(vaultDir, 0750)

	if err != nil {
		log.Fatal(err)
	}

	ti := textinput.New()
	ti.Placeholder = "Enter file name"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{})
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fff"))

	ta := textarea.New()
	ta.Placeholder = "Write your note here..."
	ta.Focus()
	ta.ShowLineNumbers = false

	return model{
		newFileInput:           ti,
		createFileInputVisible: false,
		noteTextArea:           ta,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+n":
			m.createFileInputVisible = true
			return m, nil

		case "ctrl+s":
			if m.currentFile == nil {
				break
			}

			if err := m.currentFile.Truncate(0); err != nil {
				fmt.Println("cannot save the file 🥲")
				return m, nil
			}

			if _, err := m.currentFile.Seek(0, 0); err != nil {
				fmt.Println("cannot save the file 🥲")
				return m, nil
			}

			if _, err := m.currentFile.WriteString(m.noteTextArea.Value()); err != nil {
				fmt.Println("cannot save the file 🥲")
				return m, nil
			}

			if err := m.currentFile.Close(); err != nil {
				fmt.Println("cannot close the file")
			}

			m.currentFile = nil
			m.noteTextArea.SetValue("")

			return m, nil

		case "enter":
			if m.currentFile != nil {
				break
			}

			filename := m.newFileInput.Value()

			if filename != "" {
				filepath := fmt.Sprintf("%s/%s.md", vaultDir, filename)

				if _, err := os.Stat(filepath); err == nil {
					return m, nil
				}

				file, err := os.Create(filepath)

				if err != nil {
					log.Fatalf("%v", err)
				}

				m.currentFile = file
				m.createFileInputVisible = false
				m.newFileInput.SetValue("")
			}

			return m, nil
		}

	}

	if m.createFileInputVisible {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
	}

	if m.currentFile != nil {
		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("205")).
		PaddingLeft(2).
		PaddingRight(2)

	welcome := style.Render("Welcome to totion 🧠")

	help := "Ctrl+N: new file . Ctrl+L: List . Esc: back/save . Ctrl+S: Save . Ctrl+Q: quit"

	view := ""

	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}

	if m.currentFile != nil {
		view = m.noteTextArea.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, thers's been an error: %v", err)
		os.Exit(1)
	}
}
