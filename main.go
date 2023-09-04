package main

import (
	"fmt"
	"runtime"

	"github.com/xackery/velium-server/server"
)

// icon: https://prefinem.com/simple-icon-generator/#eyJiYWNrZ3JvdW5kQ29sb3IiOiIjMDA4MGZmIiwiYm9yZGVyQ29sb3IiOiIjMDAwMDAwIiwiYm9yZGVyV2lkdGgiOiI0IiwiZXhwb3J0U2l6ZSI6IjI1NiIsImV4cG9ydGluZyI6ZmFsc2UsImZvbnRGYW1pbHkiOiJBYmhheWEgTGlicmUiLCJmb250UG9zaXRpb24iOiI1NSIsImZvbnRTaXplIjoiMjMiLCJmb250V2VpZ2h0Ijo2MDAsImltYWdlIjoiIiwiaW1hZ2VNYXNrIjoiIiwiaW1hZ2VTaXplIjo1MCwic2hhcGUiOiJkaWFtb25kIiwidGV4dCI6IlYifQ

var (
	// Version is the current version of the server
	Version = "0.0.1"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		if runtime.GOOS == "windows" {
			fmt.Println("Press Enter to exit...")
			fmt.Scanln()
		}
	}
}

func run() error {
	fmt.Println("Starting Velium Server v" + Version)
	server.Start()
	return nil
}
