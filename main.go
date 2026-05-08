package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	runs := flag.Int("runs", 15, "number of CI runs to display in the right pane")
	flag.Parse()

	p := tea.NewProgram(
		NewModel(*runs),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "github-watch: %v\n", err)
		os.Exit(1)
	}
}
