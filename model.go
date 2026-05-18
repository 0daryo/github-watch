package main

import (
	"fmt"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	prCIRefresh      = 60 * time.Second
	rateLimitRefresh = 30 * time.Second
)

// Pane represents which pane has focus.
type Pane int

const (
	PanePR Pane = iota
	PaneCI
)

// Messages
type prCITickMsg time.Time
type rateLimitTickMsg time.Time

type runsMsg struct {
	runs []Run
	err  error
}
type prsMsg struct {
	prs []PR
	err error
}
type rateLimitMsg GitHubLimits

type Model struct {
	width    int
	height   int
	pane     Pane
	runs     []Run
	prs      []PR
	prCursor int
	ciCursor int
	runsErr  error
	prsErr   error
	limits      GitHubLimits
	lastUpd     time.Time
	loadingRuns bool
	loadingPRs  bool
	maxRuns     int
}

func NewModel(maxRuns int) Model {
	return Model{loadingRuns: true, loadingPRs: true, pane: PanePR, maxRuns: maxRuns}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchPRCI(m.maxRuns),
		fetchRateLimit(),
		prCITickCmd(),
		rateLimitTickCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loadingRuns = true
			m.loadingPRs = true
			return m, tea.Batch(fetchPRCI(m.maxRuns), fetchRateLimit())
		case "tab", "h", "l", "left", "right":
			if m.pane == PanePR {
				m.pane = PaneCI
			} else {
				m.pane = PanePR
			}
		case "j", "down":
			if m.pane == PanePR {
				if m.prCursor < len(m.prs)-1 {
					m.prCursor++
				}
			} else {
				if m.ciCursor < len(m.runs)-1 {
					m.ciCursor++
				}
			}
		case "k", "up":
			if m.pane == PanePR {
				if m.prCursor > 0 {
					m.prCursor--
				}
			} else {
				if m.ciCursor > 0 {
					m.ciCursor--
				}
			}
		case "enter":
			return m, m.openSelected()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case prCITickMsg:
		m.loadingRuns = true
		m.loadingPRs = true
		return m, tea.Batch(fetchPRCI(m.maxRuns), prCITickCmd())

	case rateLimitTickMsg:
		return m, tea.Batch(fetchRateLimit(), rateLimitTickCmd())

	case runsMsg:
		m.loadingRuns = false
		m.lastUpd = time.Now()
		if msg.err != nil {
			m.runsErr = msg.err
		} else {
			m.runs = msg.runs
			m.runsErr = nil
			if m.ciCursor >= len(m.runs) && len(m.runs) > 0 {
				m.ciCursor = len(m.runs) - 1
			}
		}

	case prsMsg:
		m.loadingPRs = false
		m.lastUpd = time.Now()
		if msg.err != nil {
			m.prsErr = msg.err
		} else {
			m.prs = msg.prs
			m.prsErr = nil
			if m.prCursor >= len(m.prs) && len(m.prs) > 0 {
				m.prCursor = len(m.prs) - 1
			}
		}

	case rateLimitMsg:
		m.limits = GitHubLimits(msg)
	}

	return m, nil
}

func (m Model) isLoading() bool {
	return m.loadingRuns || m.loadingPRs
}

func (m Model) openSelected() tea.Cmd {
	if m.pane == PanePR && m.prCursor < len(m.prs) {
		pr := m.prs[m.prCursor]
		return tea.ExecProcess(
			exec.Command("gh", "pr", "view", "--web", fmt.Sprintf("%d", pr.Number)),
			func(err error) tea.Msg { return nil },
		)
	}
	if m.pane == PaneCI && m.ciCursor < len(m.runs) {
		run := m.runs[m.ciCursor]
		return tea.ExecProcess(
			exec.Command("gh", "run", "view", "--web", fmt.Sprintf("%d", run.ID)),
			func(err error) tea.Msg { return nil },
		)
	}
	return nil
}

func prCITickCmd() tea.Cmd {
	return tea.Tick(prCIRefresh, func(t time.Time) tea.Msg {
		return prCITickMsg(t)
	})
}

func rateLimitTickCmd() tea.Cmd {
	return tea.Tick(rateLimitRefresh, func(t time.Time) tea.Msg {
		return rateLimitTickMsg(t)
	})
}

func fetchPRCI(maxRuns int) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			runs, err := FetchRuns(maxRuns)
			return runsMsg{runs: runs, err: err}
		},
		func() tea.Msg {
			prs, err := FetchPRs()
			return prsMsg{prs: prs, err: err}
		},
	)
}

func fetchRateLimit() tea.Cmd {
	return func() tea.Msg {
		return rateLimitMsg(FetchGitHubLimits())
	}
}
