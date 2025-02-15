package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Check if a commit message is provided
	if len(os.Args) < 2 {
		fmt.Println("Please provide a commit message as an argument.")
		return
	}

	// Get the commit message from the command line argument
	commitMessage := strings.Join(os.Args[1:], " ")

	// Execute 'git add .'
	err := executeGitCommand("git", "add", ".")
	if err != nil {
		fmt.Println("Error executing git add:", err)
		return
	}

	// Execute 'git commit -m "{commitMessage}"'
	err = executeGitCommand("git", "commit", "-m", commitMessage)
	if err != nil {
		fmt.Println("Error executing git commit:", err)
		return
	}

	fmt.Println("Changes have been committed successfully!")
}

// executeGitCommand runs the given git command
func executeGitCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command and return any errors
	return cmd.Run()
}
