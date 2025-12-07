package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/list"
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

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	newFileInput           textinput.Model
	createFileInputVisible bool
	currentFile            *os.File
	noteTextArea           textarea.Model
	list                   list.Model
	showList               bool
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

	noteList := listFiles()
	finalList := list.New(noteList, list.NewDefaultDelegate(), 0, 0)
	finalList.Title = "All notes 📝"
	finalList.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("254")).Padding(0, 1)

	return model{
		newFileInput:           ti,
		createFileInputVisible: false,
		noteTextArea:           ta,
		list:                   finalList,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+l":
			noteList := listFiles()
			m.list.SetItems(noteList)
			m.showList = true
			return m, nil

		case "esc":
			if m.createFileInputVisible {
				m.createFileInputVisible = false
			}

			if m.currentFile != nil {
				m.noteTextArea.SetValue("")
				m.currentFile = nil
			}

			if m.showList {
				if m.list.FilterState() == list.Filtering {
					break
				}
				m.showList = false
			}

			return m, nil

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

			if m.showList {
				item, ok := m.list.SelectedItem().(item)
				if ok {
					filpath := fmt.Sprintf("%s/%s", vaultDir, item.title)
					content, err := os.ReadFile(filpath)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						return m, nil
					}
					m.noteTextArea.SetValue(string(content))
					f, err := os.OpenFile(filpath, os.O_RDWR, 0644)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						return m, nil
					}
					m.currentFile = f
					m.showList = false
				}
				return m, nil
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

	if m.showList {
		m.list, cmd = m.list.Update(msg)
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

	if m.showList {
		view = m.list.View()
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

func listFiles() []list.Item {
	items := make([]list.Item, 0)

	entries, err := os.ReadDir(vaultDir)

	if err != nil {
		log.Fatal("Error reading notes")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			modTime := info.ModTime().Format("2006-01-02 15:04")

			items = append(items, item{
				title: entry.Name(),
				desc:  fmt.Sprintf("Modified: %s", modTime),
			})
		}
	}

	return items
}
