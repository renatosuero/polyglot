package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"archive/zip"

	"github.com/labstack/gommon/log"
)

type Downloader struct {
}

func (d *Downloader) Download(url, dest string) (string, error) {
	log.Infof("Downloading %s into %s", url, dest)
	parts := strings.Split(url, "/")
	name := fmt.Sprintf("%s_%s", parts[len(parts)-2], parts[len(parts)-1])
	zipName := fmt.Sprintf("%s.zip", name)
	zipURL := fmt.Sprintf("%s/archive/master.zip", url)

	resp, err := http.Get(zipURL)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	path := filepath.Join(dest, zipName)
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	downloadedPath := filepath.Join(dest, name)
	unzip(path, downloadedPath)

	return downloadedPath, nil
}

func unzip(path, dest string) error {
	log.Infof("Unzipping %s into %s", path, dest)
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	for i := range reader.File {
		file := reader.File[i]
		zippedFile, err := file.Open()
		if err != nil {
			return err
		}

		defer zippedFile.Close()
		extractedFilePath := filepath.Join(
			dest,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			os.MkdirAll(extractedFilePath, file.Mode())
			continue
		}

		outputFile, err := os.OpenFile(
			extractedFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			file.Mode(),
		)
		if err != nil {
			return err
		}

		_, err = io.Copy(outputFile, zippedFile)
		if err != nil {
			return err
		}
	}
	os.Remove(path)

	return nil
}
