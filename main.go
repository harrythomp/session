package main

import (
	"fmt"
	"harry/session/src/config"
	"harry/session/src/session"
)

func main() {
	sessions, err := session.FindSessions(config.Config{
		SearchPaths: []string{
			"~/dev/*",
			"~/work/*",
		},
		IncludePaths: []string{
			"~/env",
		},
	})
	if err != nil {
		panic(err)
	}

	for _, session := range sessions {
		fmt.Printf("%v\n", session)
	}
}
