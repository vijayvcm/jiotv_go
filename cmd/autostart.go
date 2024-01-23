package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"runtime"
)

// AutoStart adds or updates an auto start command for the current
// executable in the bashrc file. It checks if auto start already
// exists, gets user consent if needed, and adds the command
// to run the executable on startup, passing any extra args.
// It supports Linux, Android and OSX systems.
func AutoStart(extraArgs string) error {
	if runtime.GOOS == "linux" || runtime.GOOS == "android" || runtime.GOOS == "darwin" {
		// Get the path to the current binary
		selfPath, err := os.Executable()
		if err != nil {
			return err
		}

		var bashrcPath string

		// Check if it's a Termux system
		isTermux := isTermux()
		if isTermux {
			// For Termux, use the system-wide bashrc
			bashrcPath = os.Getenv("PREFIX") + "/etc/bash.bashrc"
		} else {
			// For Linux, use the user-specific bashrc
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			bashrcPath = userHomeDir + "/.bashrc"

			// Check if the bashrc file exists
			if _, err := os.Stat(bashrcPath); os.IsNotExist(err) {
				//  ask consent to create the bashrc file
				fmt.Printf("Make sure you have\nBashrc file not found at %s. Would you like to create it? (y/n): ", bashrcPath)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) == "y" {
					// Create the bashrc file
					_, err := os.Create(bashrcPath)
					if err != nil {
						return err
					}
				} else {
					fmt.Println("Auto start canceled by user.")
					return nil
				}
			}
		}

		// Check if the auto start line is already present
		autoStartLine := fmt.Sprintf("%s run", selfPath)
		exists, err := grep(bashrcPath, autoStartLine)
		if err != nil {
			return err
		}

		if !exists {
			// Get user consent
			consent := getConsentFromUser()
			if !consent {
				fmt.Println("Auto start canceled by user.")
				return nil
			}
			fmt.Printf("Adding auto start to %s...\n", bashrcPath)
			err := addToBashrc(bashrcPath, autoStartLine+" "+extraArgs)
			if err != nil {
				return err
			}
		} else {
			fmt.Printf("Removing existing auto start from %s...\n", bashrcPath)
			err := removeFromBashrc(bashrcPath, autoStartLine)
			if err != nil {
				return err
			}
			// Add the auto start line with extra args
			err = addToBashrc(bashrcPath, autoStartLine+" "+extraArgs)
			if err != nil {
				return err
			}
		}

		return nil
	}
	fmt.Printf("Auto start is not supported on %s\n", runtime.GOOS)
	return nil
}

// isTermux checks if the environment variable PREFIX is set, which indicates
// the program is running in Termux on Android. Returns true if running in
// Termux, false otherwise.
func isTermux() bool {
	termuxProperty := os.Getenv("PREFIX")
	return termuxProperty != ""
}

// getConsentFromUser prompts the user to consent to auto start and returns
// a boolean indicating if consent was given. If running in Termux, consent
// is assumed. Otherwise, the user is prompted in the terminal.
func getConsentFromUser() bool {
	if isTermux() {
		return true
	}
	fmt.Print("Warning: Auto start may not be suitable for all systems. We only support BASH Terminal. Do you consent to continue? (y/n): ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

// grep searches the given filename for lines containing the pattern string.
// It returns a bool indicating if the pattern was found, and an error if one occurred while reading the file.
func grep(filename, pattern string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			return true, nil
		}
	}

	return false, file.Close()
}

// addToBashrc appends the given line to the specified bashrc file.
// It opens the file in append mode, writes the line, and closes the file.
// Returns any error encountered.
func addToBashrc(filename, line string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(file, line)
	if err != nil {
		return err
	}

	return file.Close()
}

// removeFromBashrc removes the given line from the specified bashrc file.
// It opens the file, scans each line to build a new slice excluding the given line,
// closes and deletes the original file, recreates it, writes the new lines slice,
// and closes the new file. Returns any error encountered.
func removeFromBashrc(filename, line string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		currentLine := scanner.Text()
		if !strings.Contains(currentLine, line) {
			lines = append(lines, currentLine)
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}

	err = os.Remove(filename)
	if err != nil {
		return err
	}

	newFile, err := os.Create(filename)
	if err != nil {
		return err
	}

	for _, l := range lines {
		_, err = fmt.Fprintln(newFile, l)
		if err != nil {
			return err
		}
	}

	return newFile.Close()
}
