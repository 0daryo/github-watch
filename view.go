package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

const barWidth = 20

func (m Model) View() string {
	var b strings.Builder
	w := m.effectiveWidth()

	// Title bar
	refresh := "15s"
	if m.isLoading() {
		refresh = "..."
	}
	b.WriteString(titleBarStyle.Width(w).Render(
		fmt.Sprintf("github-watch    Refresh: %s    [tab/h/l]pane [j/k]nav [enter]open [r]efresh [q]uit", refresh),
	))
	b.WriteString("\n")

	// Calculate pane dimensions
	leftW := w/2 - 1
	rightW := w - leftW - 1
	topH := m.topPaneHeight()

	leftPane := m.renderPRPane(leftW, topH)
	rightPane := m.renderCIPane(rightW, topH)

	joined := lipgloss.JoinHorizontal(lipgloss.Top,
		leftPane,
		lipgloss.NewStyle().Foreground(colorMuted).Render(strings.TrimRight(strings.Repeat("│\n", topH), "\n")),
		rightPane,
	)
	b.WriteString(joined)

	// Separator line
	b.WriteString(dimStyle.Render(strings.Repeat("─", w)))
	b.WriteString("\n")

	// Bottom pane: rate limits
	b.WriteString(m.renderRateLimitPane(w))

	// Footer
	footer := dimStyle.Render("  Focus: ")
	if m.pane == PanePR {
		footer += selectedStyle.Render(" PRs ") + dimStyle.Render("  CI Runs")
	} else {
		footer += dimStyle.Render("  PRs  ") + selectedStyle.Render(" CI Runs ")
	}
	if !m.lastUpd.IsZero() {
		footer += dimStyle.Render("     Updated: " + m.lastUpd.Format("15:04:05"))
	}
	b.WriteString(footer + "\n")

	return b.String()
}

func (m Model) renderPRPane(w, maxLines int) string {
	var lines []string

	// Header
	title := " Pull Requests"
	if m.pane == PanePR {
		lines = append(lines, sectionStyle.Render(title))
	} else {
		lines = append(lines, dimStyle.Render(title))
	}
	lines = append(lines, dimStyle.Render(" "+strings.Repeat("─", w-2)))

	if m.prsErr != nil {
		lines = append(lines, " "+errorStyle.Render(fmt.Sprintf("Error: %v", m.prsErr)))
		return padLines(lines, w, maxLines)
	}
	if len(m.prs) == 0 {
		lines = append(lines, " "+mutedStyle.Render("No open PRs"))
		return padLines(lines, w, maxLines)
	}

	maxTitle := w - 30
	if maxTitle < 10 {
		maxTitle = 10
	}

	for i, pr := range m.prs {
		revIcon := ReviewIcon(pr.ReviewDecision)
		revColor := ReviewColor(pr.ReviewDecision)

		passed, failed, pending := ChecksSummary(pr)
		checksStr := fmt.Sprintf("✓%d✗%d◷%d", passed, failed, pending)
		var checksColor string
		switch {
		case failed > 0:
			checksColor = string(colorRed)
		case pending > 0:
			checksColor = string(colorYellow)
		default:
			checksColor = string(colorGreen)
		}

		title := truncate(pr.Title, maxTitle)
		num := fmt.Sprintf("#%d", pr.Number)

		line := fmt.Sprintf(" %-5s %-*s %s %s",
			num,
			maxTitle, title,
			lipgloss.NewStyle().Foreground(lipgloss.Color(revColor)).Width(12).Render(revIcon),
			lipgloss.NewStyle().Foreground(lipgloss.Color(checksColor)).Render(checksStr),
		)

		if i == m.prCursor && m.pane == PanePR {
			lines = append(lines, selectedStyle.Width(w).Render(line))
		} else {
			lines = append(lines, line)
		}
	}

	lines = append(lines, dimStyle.Render(fmt.Sprintf(" %d open", len(m.prs))))

	return padLines(lines, w, maxLines)
}

func (m Model) renderCIPane(w, maxLines int) string {
	var lines []string

	title := " CI Runs"
	if m.pane == PaneCI {
		lines = append(lines, sectionStyle.Render(title))
	} else {
		lines = append(lines, dimStyle.Render(title))
	}
	lines = append(lines, dimStyle.Render(" "+strings.Repeat("─", w-2)))

	if m.runsErr != nil {
		lines = append(lines, " "+errorStyle.Render(fmt.Sprintf("Error: %v", m.runsErr)))
		return padLines(lines, w, maxLines)
	}
	if len(m.runs) == 0 {
		lines = append(lines, " "+mutedStyle.Render("No runs"))
		return padLines(lines, w, maxLines)
	}

	nameW := 16
	branchW := w - nameW - 18
	if branchW < 8 {
		branchW = 8
	}

	for i, r := range m.runs {
		icon := RunStatusIcon(r)
		color := RunStatusColor(r)

		statusStr := r.Status
		if r.Conclusion != "" {
			statusStr = r.Conclusion
		}

		line := fmt.Sprintf(" %s %-*s %-*s %-10s %s",
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(icon),
			nameW, truncate(r.Name, nameW),
			branchW, branchStyle.Render(truncate(r.Branch, branchW)),
			lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(truncate(statusStr, 10)),
			dimStyle.Render(RunElapsed(r)),
		)

		if i == m.ciCursor && m.pane == PaneCI {
			lines = append(lines, selectedStyle.Width(w).Render(line))
		} else {
			lines = append(lines, line)
		}
	}

	// Summary
	inProgress, success, failed := 0, 0, 0
	for _, r := range m.runs {
		switch {
		case r.Status == "in_progress" || r.Status == "queued" || r.Status == "waiting":
			inProgress++
		case r.Conclusion == "success":
			success++
		case r.Conclusion == "failure":
			failed++
		}
	}
	lines = append(lines, fmt.Sprintf(" %s%d %s%d %s%d",
		lipgloss.NewStyle().Foreground(colorYellow).Render("⟳"), inProgress,
		lipgloss.NewStyle().Foreground(colorGreen).Render("✓"), success,
		lipgloss.NewStyle().Foreground(colorRed).Render("✗"), failed,
	))

	return padLines(lines, w, maxLines)
}

func (m Model) renderRateLimitPane(w int) string {
	var b strings.Builder

	b.WriteString(sectionStyle.Render(" GitHub API Rate Limits"))
	b.WriteString("\n")

	if m.limits.Err != nil {
		b.WriteString("  " + errorStyle.Render(fmt.Sprintf("Error: %v", m.limits.Err)))
		b.WriteString("\n")
		return b.String()
	}

	// Show all three limits in a single compact row layout
	b.WriteString(renderRateLimitRow("Core", m.limits.Core.Remaining, m.limits.Core.Limit))
	b.WriteString(renderRateLimitRow("GraphQL", m.limits.GraphQL.Remaining, m.limits.GraphQL.Limit))
	b.WriteString(renderRateLimitRow("Search", m.limits.Search.Remaining, m.limits.Search.Limit))
	b.WriteString(renderResetRow(m.limits.Core.Reset))

	return b.String()
}

func renderRateLimitRow(label string, remaining, limit int) string {
	if limit == 0 {
		return fmt.Sprintf("  %s %s\n", labelStyle.Render("  "+label), dimStyle.Render("n/a"))
	}
	pct := float64(remaining) / float64(limit) * 100
	bar := progressBar(pct)
	return fmt.Sprintf("  %s %s %5d/%-5d %3.0f%%\n",
		labelStyle.Render("  "+label), bar, remaining, limit, pct)
}

func renderResetRow(resetTime time.Time) string {
	if resetTime.IsZero() {
		return ""
	}
	dur := time.Until(resetTime)
	if dur < 0 {
		dur = 0
	}
	var resetStr string
	if dur > time.Minute {
		resetStr = fmt.Sprintf("%dm", int(dur.Minutes()))
	} else {
		resetStr = fmt.Sprintf("%ds", int(dur.Seconds()))
	}
	return fmt.Sprintf("  %s %s\n",
		labelStyle.Render("  Resets in:"), dimStyle.Render(resetStr))
}

func progressBar(pct float64) string {
	filled := int(pct / 100 * barWidth)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	empty := barWidth - filled
	color := barColor(pct)

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(colorMuted)

	return filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))
}

// ── Layout helpers ──────────────────────────────────────────────────────────

func padLines(lines []string, w, maxLines int) string {
	for len(lines) < maxLines {
		lines = append(lines, "")
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

func (m Model) topPaneHeight() int {
	// Reserve lines for: title bar(1) + separator(1) + rate limit(~6) + footer(1)
	reserved := 9
	h := m.height - reserved
	if h < 10 {
		h = 10
	}
	return h
}

func (m Model) effectiveWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 120
}

func truncate(s string, maxLen int) string {
	if maxLen < 1 {
		maxLen = 1
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-1]) + "…"
}
