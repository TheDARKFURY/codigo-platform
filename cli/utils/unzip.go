package utils

import (
	"archive/zip"
	"bytes"
	"codigo/cli/sentry"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Writer func(projectRoot string, file *zip.File, fileData *bytes.Buffer) error

func Unzip(reader *zip.Reader, projectRoot string, writer Writer) error {
	for _, file := range reader.File {
		err := func() error {
			Log(fmt.Sprintf("Creating %s...", file.Name))

			dirPath := ""

			if strings.Contains(file.Name, "/") {
				split := strings.Split(file.Name, "/")
				dirPath = strings.Join(split[:len(split)-1], "/")
			}

			if err := os.MkdirAll(path.Join(projectRoot, dirPath), 0755); err != nil {
				return err
			}

			reader, err := file.Open()

			if err != nil {
				return err
			}

			defer func(reader io.ReadCloser) {
				err := reader.Close()
				if err != nil {
					sentry.ReportGenericError(err)
				}
			}(reader)

			buffer := new(bytes.Buffer)

			_, err = io.CopyN(buffer, reader, file.FileInfo().Size())

			if err != nil {
				return err
			}

			return writer(projectRoot, file, buffer)
		}()

		if err != nil {
			return err
		}
	}

	return nil
}
