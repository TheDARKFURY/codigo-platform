package utils

import (
	"archive/zip"
	"bytes"
	"codigo/cli/sentry"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func AskUserToContinue() error {
	Warning("Unable to apply the file diff - \"git\" command not found.\n\nThe implemented code will be lost.\nWould you like to proceed? (y/n)")

	var answer string

	if _, err := fmt.Scanln(&answer); err != nil {
		return err
	}

	answer = strings.ToLower(answer)

	if answer != "y" && answer != "yes" {
		return fmt.Errorf("the user has cancelled the generation process")
	}

	return nil
}

func DiffFileUsingGit(projectRoot string, file *zip.File, fileData *bytes.Buffer) error {
	// Create a diff directory under .codigo
	diffPath := path.Join(projectRoot, ".codigo", "diff")

	if err := os.MkdirAll(diffPath, 0755); err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(diffPath); err != nil {
			sentry.ReportGenericError(err)
		}
	}()

	// Create any directory where this file must exist
	dirPath := ""

	if strings.Contains(file.Name, "/") {
		split := strings.Split(file.Name, "/")
		dirPath = strings.Join(split[:len(split)-1], "/")
	}

	if err := os.MkdirAll(path.Join(diffPath, dirPath), 0755); err != nil {
		return err
	}

	// Create an auxiliary file require for by the git merge-file command
	auxFilePath := path.Join(diffPath, "aux")
	auxFile, err := os.OpenFile(auxFilePath, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	// Close the aux file immediately because git will be using it
	if err := auxFile.Close(); err != nil {
		return err
	}

	// Write the generated file in the diff dir
	generatedFilePath := path.Join(diffPath, file.Name)

	if err := os.WriteFile(generatedFilePath, fileData.Bytes(), 0644); err != nil {
		return err
	}

	currentFilePath := path.Join(projectRoot, file.Name)

	// Check if file contains pending diff to be resolved
	currentFile, err := os.ReadFile(currentFilePath)

	if err != nil {
		return err
	}

	if strings.Contains(string(currentFile), "<<<<<<<") || strings.Contains(string(currentFile), "=======") {
		return fmt.Errorf("the file %s has pending diff conflicts to be resolved", currentFilePath)
	}

	// Generate file diff
	diff, err := exec.Command(
		"git",
		"merge-file",
		"-p",
		"-L",
		"current",
		"-L",
		"x",
		"-L",
		"new",
		currentFilePath,
		auxFilePath,
		generatedFilePath,
	).CombinedOutput()

	if strings.Contains(string(diff), "<<<<<<<") {
		if err := os.WriteFile(currentFilePath, diff, 0644); err != nil {
			return err
		}

		Warning(fmt.Sprintf("Action required: attend diff conflicts at\n%s", currentFilePath))
		return nil
	}

	return err
}
