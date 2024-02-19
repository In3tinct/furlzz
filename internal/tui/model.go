package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nsecho/furlzz/mutator"
	"strconv"
	"strings"
	"time"
)

type Model struct {
	Crash       bool
	Runs        uint
	Timeout     uint
	App         string
	Device      string
	Function    string
	Method      string
	Delegate    string
	UIApp       string
	Scene       string
	Base        string
	Input       string
	ValidInputs []string

	exiting    bool
	start      time.Time
	seconds    int
	ctr        int
	op         string
	ur         string
	lastInfo   string
	messages   []string
	errorTimes []string
	lastErr    string
}

type StatsMsg string
type ErrMsg string
type MutatedMsg *mutator.Mutated
type tickMsg time.Time
type SessionDetached struct{}

func (m Model) Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func NewModel() Model {
	m := Model{}

	m.seconds = 5
	m.start = time.Now()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.exiting = true
			return m, m.Tick()
		}
	case StatsMsg:
		ms := fmt.Sprintf("+%ds=>%s", int(time.Since(m.start).Seconds()), string(msg))
		m.messages = append(m.messages, ms)
		return m, nil
	case ErrMsg:
		m.lastErr = fmt.Sprintf("+%ds=>%s", int(time.Since(m.start).Seconds()), string(msg))
		m.exiting = true
		return m, m.Tick()
	case MutatedMsg:
		m.ctr++
		m.op = msg.Mutation
		m.ur = msg.Input
	case tickMsg:
		m.seconds--
		if m.seconds <= 0 {
			return m, tea.Quit
		}
		return m, m.Tick()
	case SessionDetached:
		m.exiting = true
		return m, m.Tick()
	}

	return m, nil
}

func (m Model) View() string {
	s := ""
	box1 := leftBoxContainer.Render(renderBox(leftBox,
		stripOrNA(m.App, 37),
		stripOrNA(m.Function, 37),
		stripOrNA(m.Delegate, 37),
		stripOrNA(m.Method, 37),
		stripOrNA(m.UIApp, 37),
		stripOrNA(m.Scene, 37),
		stripOrNA(m.Base, 37)))

	totalRuns := ""
	if m.Runs == 0 {
		totalRuns = "infinitely"
	} else {
		totalRuns = strconv.Itoa(int(m.Runs))
	}

	box2 := rightBoxContainer.Render(renderBox(middleBox, m.ctr, totalRuns, m.Timeout, stripOrNA(m.op, 47), stripOrNA(m.ur, 47)))

	s += lipgloss.JoinHorizontal(lipgloss.Top, box1, box2)

	// Add messages
	message := ""
	message += " " + ttlStyle.Render("Messages")
	message += "\n"

	// Error message should come on top
	errMsg := ""

	if m.lastErr != "" {
		splitted := strings.Split(m.lastErr, "=>")
		errMsg += errStyle.Render(splitted[0] + " " + splitted[1])
		errMsg += "\n"
	}

	// Add previous messages in reversed order
	reversed := make([]string, len(m.messages))
	for i := len(m.messages) - 1; i >= 0; i-- {
		if len(m.messages[i]) > 60 {
			reversed[len(m.messages)-1-i] = m.messages[i][:57] + "..."
		} else {
			reversed[len(m.messages)-1-i] = m.messages[i]
		}
		// Style the time and message
		splitted := strings.Split(reversed[len(m.messages)-1-i], "=>")
		reversed[len(m.messages)-1-i] = dataStyle.Render(splitted[0] + " " + splitted[1])
	}
	message += messageStyle.Render(errMsg + strings.Join(reversed, "\n"))

	// If we are exiting append the message in how many seconds
	if m.exiting {
		message += "\n" + " " + ttlStyle.Render(fmt.Sprintf("Exiting in %d seconds...", m.seconds))
	}
	return lipgloss.JoinVertical(lipgloss.Top, s, message)
}

func stripOrNA(s string, count int) string {
	if s == "" {
		return "N/A"
	}
	if len(s) > count {
		return s[:count-3] + "..."
	}
	return s
}
