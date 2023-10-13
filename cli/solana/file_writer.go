package solana

import (
	"archive/zip"
	"bytes"
	"codigo/cli/utils"
	"errors"
	"os"
	"path"
	"regexp"
	"strings"
)

func WriteFile(projectRoot string, file *zip.File, fileData *bytes.Buffer) error {
	return os.WriteFile(path.Join(projectRoot, file.Name), fileData.Bytes(), 0644)
}

func WriteFileForNative(projectRoot string, file *zip.File, fileData *bytes.Buffer) error {
	diffFile := false

	if file.Name == "Cargo.toml" {
		diffFile = true
	}

	if strings.HasPrefix(file.Name, "src/") && file.Name != "src/mod.rs" {
		diffFile = true
	}

	return writeFillWithDiff(diffFile, projectRoot, file, fileData)
}

func WriteFileForAnchor(projectRoot string, file *zip.File, fileData *bytes.Buffer) error {
	diffFile := false

	if file.Name == "Cargo.toml" {
		diffFile = true
	}

	if file.Name == "package.json" {
		diffFile = true
	}

	if strings.HasPrefix(file.Name, "tests/") {
		diffFile = true
	}

	regex := regexp.MustCompile(`^programs/([^/]+)/src/stubs/([^/]+)$`)

	if regex.MatchString(file.Name) && !strings.HasSuffix(file.Name, "mod.rs") {
		diffFile = true
	}

	regex = regexp.MustCompile(`^programs/([^/]+)/Cargo.toml$`)

	if regex.MatchString(file.Name) {
		diffFile = true
	}

	regex = regexp.MustCompile(`^programs/([^/]+)/Xargo.toml$`)

	if regex.MatchString(file.Name) {
		diffFile = true
	}

	return writeFillWithDiff(diffFile, projectRoot, file, fileData)
}

func writeFillWithDiff(diff bool, projectRoot string, file *zip.File, fileData *bytes.Buffer) error {
	if diff && utils.IsGitInstalled() {
		// If file doesn't exist, just create it
		_, err := os.Stat(path.Join(projectRoot, file.Name))

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return WriteFile(projectRoot, file, fileData)
			}

			return err
		}

		return utils.DiffFileUsingGit(projectRoot, file, fileData)
	}

	return WriteFile(projectRoot, file, fileData)
}
