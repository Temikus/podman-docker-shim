package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var debug bool

func init() {
	debug = os.Getenv("PODMAN_SHIM_DEBUG") != ""
}

func debugLog(format string, v ...interface{}) {
	if debug {
		log.Printf("DEBUG: "+format, v...)
	}
}

func checkPodmanInstalled() bool {
	path, err := exec.LookPath("podman")
	if debug {
		if err != nil {
			debugLog("podman not found in PATH")
		} else {
			debugLog("found podman at: %s", path)
		}
	}
	return err == nil
}

func getAllTags(image string) ([]string, error) {
	debugLog("getting all tags for image: %s", image)
	cmd := exec.Command("podman", "images", "--filter", "reference="+image, "--format", "{{.Tag}}")
	debugLog("executing command: %v", cmd.Args)

	out, err := cmd.Output()
	if err != nil {
		debugLog("error getting tags: %v", err)
		return nil, fmt.Errorf("error getting tags: %w", err)
	}

	var tags []string
	for _, tag := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if tag != "" && tag != "<none>" {
			tags = append(tags, tag)
		}
	}

	debugLog("found %d tags: %v", len(tags), tags)

	if len(tags) == 0 {
		return nil, fmt.Errorf("no tags found for image: %s", image)
	}

	return tags, nil
}

func pushAllTags(image string) error {
	debugLog("pushing all tags for image: %s", image)
	tags, err := getAllTags(image)
	if err != nil {
		return err
	}

	log.Printf("Found tags for %s:\n", image)
	for _, tag := range tags {
		log.Printf("  %s\n", tag)
	}
	log.Println("\nStarting push operation...")

	successful := 0
	failed := 0

	for _, tag := range tags {
		fullTag := fmt.Sprintf("%s:%s", image, tag)
		log.Printf("Pushing %s...\n", fullTag)

		cmd := exec.Command("podman", "push", fullTag)
		debugLog("executing command: %v", cmd.Args)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			failed++
			debugLog("failed to push %s: %v", fullTag, err)
			log.Printf("Failed to push %s: %v\n", fullTag, err)
		} else {
			successful++
			debugLog("successfully pushed %s", fullTag)
			log.Printf("Successfully pushed %s\n", fullTag)
		}
	}

	log.Printf("\nPush completed: %d successful, %d failed\n", successful, failed)
	if failed > 0 {
		return fmt.Errorf("%d pushes failed", failed)
	}
	return nil
}

func executeCommand(args []string) error {
	cmd := exec.Command("podman", args...)
	debugLog("executing command: %v", cmd.Args)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	// prefix all log messages with "podman-shim: "
	log.SetPrefix("\x1b[93mpodman-shim: \x1b[0m") // yellow
	// disable timestamps in log messages
	log.SetFlags(0)

	debugLog("starting podman-shim")
	debugLog("args: %v", os.Args)

	if !checkPodmanInstalled() {
		log.Fatal("podman not found")
	}

	args := os.Args[1:]
	if len(args) == 0 {
		debugLog("no arguments provided, executing podman directly")
		if err := executeCommand(args); err != nil {
			debugLog("podman execution failed: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Get the command (first argument)
	command := args[0]
	debugLog("command: %s", command)

	// Only process special flags for the push command
	if command == "push" {
		args = args[1:] // Remove "push" as we'll add it back later
		debugLog("processing push command")

		// Check for -a or --all-tags flag
		hasAllTags := false
		var filteredArgs []string
		var image string

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if arg == "-a" || arg == "--all-tags" {
				hasAllTags = true
				debugLog("found all-tags flag")
			} else {
				if !strings.HasPrefix(arg, "-") {
					image = arg // Last non-flag argument is the image
					debugLog("found image argument: %s", image)
				}
				filteredArgs = append(filteredArgs, arg)
			}
		}

		if hasAllTags {
			log.Println("Unsupported flag detected for push command, invoking shim...")
			if image == "" {
				log.Fatal("no image specified")
			}
			log.Printf("Pushing all tags for image: %s", image)
			if err := pushAllTags(image); err != nil {
				log.Fatal(err)
			}
			return
		}

		// Regular push command
		args = append([]string{"push"}, filteredArgs...)
	}

	// Execute podman with the original arguments
	if err := executeCommand(args); err != nil {
		debugLog("podman execution failed: %v", err)
		os.Exit(1)
	}
}
