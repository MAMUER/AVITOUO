package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CreatePhotoZip(sourceDir string, zipPath string) (string, error) {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := zipFile.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	zipWriter := zip.NewWriter(zipFile)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var fileNames []string
	fileIndex := 1

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() || !isImage(entry.Name()) {
			continue
		}

		srcPath := filepath.Join(sourceDir, entry.Name())
		newName := fmt.Sprintf("%s_%d%s", filepath.Base(sourceDir), fileIndex, filepath.Ext(entry.Name()))
		fileNames = append(fileNames, newName)

		fileToZip, openErr := os.Open(srcPath)
		if openErr != nil {
			return "", openErr
		}

		w, createErr := zipWriter.Create(newName)
		if createErr != nil {
			closeErr := fileToZip.Close()
			if closeErr != nil {
				return "", fmt.Errorf("zip create error: %w; close error: %v", createErr, closeErr)
			}
			return "", createErr
		}

		if _, copyErr := io.Copy(w, fileToZip); copyErr != nil {
			closeErr := fileToZip.Close()
			if closeErr != nil {
				return "", fmt.Errorf("copy error: %w; close error: %v", copyErr, closeErr)
			}
			return "", copyErr
		}
		if closeErr := fileToZip.Close(); closeErr != nil {
			return "", closeErr
		}
		fileIndex++

		if fileIndex > 10 {
			break
		}
	}

	return strings.Join(fileNames, "|"), nil
}

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func CheckTotalSize(zipPath, xlsxPath string) error {
	const maxSize = 100 * 1024 * 1024
	var totalSize int64
	if info, err := os.Stat(zipPath); err == nil {
		totalSize += info.Size()
	}
	if info, err := os.Stat(xlsxPath); err == nil {
		totalSize += info.Size()
	}
	if totalSize > maxSize {
		return fmt.Errorf("общий размер файлов (%.2f МБ) превышает лимит в 100 МБ", float64(totalSize)/(1024*1024))
	}
	return nil
}
