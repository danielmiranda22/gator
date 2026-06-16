package ui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type TUIPost struct {
	ID          uuid.UUID
	Title       string
	FeedName    string
	URL         string
	Description string
	PublishedAt time.Time
}

type postItem struct {
	post TUIPost
}

func (i postItem) Title() string {
	return i.post.Title
}

func (i postItem) Description() string {
	return fmt.Sprintf("%s • %s", i.post.FeedName, i.post.PublishedAt.Format("02 Jan 2006"))
}

func (i postItem) FilterValue() string {
	return i.post.Title + " " + i.post.FeedName + " " + i.post.Description
}

type tuiModel struct {
	list        list.Model
	showPreview bool
	errMsg      string
}

type browserOpenResultMsg struct {
	err error
}

func newTUIModel(posts []TUIPost) tuiModel {
	items := make([]list.Item, 0, len(posts))
	for _, p := range posts {
		items = append(items, postItem{post: p})
	}

	l := list.New(items, list.NewDefaultDelegate(), 80, 20)
	l.Title = "Gator Posts"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	return tuiModel{
		list:        l,
		showPreview: true,
	}
}

func (m tuiModel) Init() tea.Cmd {
	log.Default().Println("Init called.")
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case browserOpenResultMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("failed to open browser: %v", msg.err)
		} else {
			m.errMsg = "opened in browser"
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "v":
			m.showPreview = !m.showPreview
			return m, nil
		case "enter":
			if item, ok := m.list.SelectedItem().(postItem); ok {
				return m, openBrowserCmd(item.post.URL)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m tuiModel) View() string {
	out := m.list.View()

	if m.showPreview {
		if item, ok := m.list.SelectedItem().(postItem); ok {
			out += "\n\n"
			out += "──────────────── Preview ────────────────\n"
			out += fmt.Sprintf("Title: %s\n", item.post.Title)
			out += fmt.Sprintf("Feed: %s\n", item.post.FeedName)
			out += fmt.Sprintf("Date: %s\n", item.post.PublishedAt.Format("02 Jan 2006 15:04"))
			out += fmt.Sprintf("ID:   %s\n", item.post.ID)
			out += fmt.Sprintf("URL:  %s\n", item.post.URL)
			if item.post.Description != "" {
				out += fmt.Sprintf("\n%s\n", item.post.Description)
			}
			out += "\n[q quit] [enter open in browser] [v toggle preview]"
		}
	} else {
		out += "\n\n[q quit] [enter open in browser] [v toggle preview]"
	}

	return out
}

func RunTUI(posts []TUIPost) error {
	p := tea.NewProgram(newTUIModel(posts))
	_, err := p.Run()
	return err
}

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		err := openBrowser(url)
		return browserOpenResultMsg{err: err}
	}
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		if isWSL() {
			cmd = exec.Command("cmd.exe", "/C", "start", url)
		} else {
			cmd = exec.Command("xdg-open", url)
		}
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}
