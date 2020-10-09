package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	path := os.Getenv("BLOCK_LOG")
	if path != "" {
		f, err := tea.LogToFile(path, "block")
		if err != nil {
			fmt.Printf("Could not open file %s: %v", path, err)
			os.Exit(1)
		}
		defer f.Close()
		log.Println("------ Starting...")
	}

	p := tea.NewProgramWithBlocks(initialize, update, view)
	if err := p.Start(); err != nil {
		fmt.Println("error starting program:", err)
		os.Exit(1)
	}
}

type model struct{}

func initialize() (tea.Model, tea.Cmd) {
	return model{}, nil
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, ok := mdl.(model)
	if !ok {
		panic("could not perform assertion on model")
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, nil
}

func view(mdl tea.Model) []tea.Block {
	_, ok := mdl.(model)
	if !ok {
		panic("could not perform assertion on model")
	}

	return []tea.Block{
		{X: 1, Y: 1, Content: "01 Hi!\n01 Meow"},
		{X: 4, Y: 4, Content: "02 Hello!\n02 Purr"},
	}
}
