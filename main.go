package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	msg string
}

func initialModel() model {
	return model{
		msg: "Hello, World",
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		}
	}
	return m, nil
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

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, thers's been an error: %v", err)
		os.Exit(1)
	}
}
