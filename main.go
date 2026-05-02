package main

import (
	"fmt"
	"harry/session/src/config"
	"harry/session/src/session"
	"harry/session/src/ui"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Error when getting user config dir: %v\n", err)
		os.Exit(1)
		return
	}

	configDir := filepath.Join(userConfigDir, "session")

	conf, err := config.ParseFromConfigDir(configDir)
	if err != nil {
		fmt.Printf("Error when parsing config: %v\n", err)
		os.Exit(1)
		return
	}

	sessions, err := session.FindSessions(conf)
	if err != nil {
		fmt.Printf("Error when finding sessions: %v\n", err)
		os.Exit(1)
		return
	}

	program := tea.NewProgram(ui.InitialModel(sessions))
	model, err := program.Run()
	if err != nil {
		fmt.Printf("Error when running program: %v\n", err)
		os.Exit(1)
		return
	}

	uiModel, ok := model.(ui.Model)
	if !ok {
		fmt.Printf("Error when casting model to ui.Model: %v\n", err)
		os.Exit(1)
		return
	}

	if uiModel.SelectedSession == nil {
		return
	}

	err = session.AttachToSession(conf, *uiModel.SelectedSession)
	if err != nil {
		fmt.Printf("Error when attaching to session: %v\n", err)
		os.Exit(1)
		return
	}
}
