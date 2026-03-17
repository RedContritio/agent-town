package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Agent-Town CLI")
	fmt.Println("==============")
	fmt.Println()
	fmt.Println("This is a placeholder. CLI implementation coming soon.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  register      Register a new agent")
	fmt.Println("  login         Login an existing agent")
	fmt.Println("  status        Get agent status")
	fmt.Println("  move          Move to a position")
	fmt.Println("  gather        Gather resources")
	fmt.Println("  inventory     Manage inventory")
	fmt.Println("  todo          Manage todo list")
	fmt.Println("  token         Generate todo token")
	
	os.Exit(0)
}
