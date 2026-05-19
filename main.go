package main

import (
	"errors"
	"fmt"
	"harry/session/src/config"
	"harry/session/src/session"
	"harry/session/src/ui"
	"io/fs"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
)

func main() {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
		return
	}

	configDir := filepath.Join(userConfigDir, "session")

	conf, err := config.ParseFromConfigDir(configDir)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
		return
	}

	sessions, err := session.FindSessions([]session.SessionFinder{
		session.PathSessionFinder{
			SearchPaths:  conf.SearchPaths,
			IncludePaths: conf.IncludePaths,
		},
		session.TmuxSessionFinder{},
	}, session.MergeSessionsPreferActive)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
		return
	}

	var selectedSession *session.Session
	if len(os.Args) > 1 {
		selectedSession, err = interpretSessionFromArg(sessions, os.Args[1])
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	} else {
		selectedSession, err = selectSession(sessions)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
	}

	if selectedSession == nil {
		return
	}

	err = session.AttachTmuxToSession(conf.Location, *selectedSession)
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
		return
	}
}

func selectSession(sessions []session.Session) (*session.Session, error) {
	program := tea.NewProgram(ui.InitialModel(sessions))
	model, err := program.Run()
	if err != nil {
		return nil, err
	}

	uiModel, ok := model.(ui.Model)
	if !ok {
		return nil, fmt.Errorf("Error when casting model to ui.Model: %v", err)
	}

	return uiModel.SelectedSession, nil
}

func interpretSessionFromArg(sessions []session.Session, arg string) (*session.Session, error) {
	argAsPath, err := filepath.Abs(arg)
	if err != nil {
		return nil, err
	}

	// Check if arg matches name or working path from regular session search
	for _, session := range sessions {
		if session.Name == arg || session.WorkingPath == argAsPath {
			return &session, nil
		}
	}

	// Check if arg is a path
	fi, err := os.Lstat(argAsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("error when reading path %s: %v", arg, err)
	}
	if err == nil && fi.IsDir() {
		// Create a new session from the path provided by arg
		s := session.NewSessionFromWorkingPath(argAsPath)
		return &s, nil
	} else {
		// Create a new session in the current working directory with the name provided by path
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		s := session.NewSessionFromWorkingPath(wd)
		s.SetName(arg)
		return &s, nil
	}
}
