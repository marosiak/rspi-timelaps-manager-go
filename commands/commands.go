package commands

import (
	"fmt"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type CommendsService struct {
	cfg *config.Config
}

func NewCommendsService(cfg *config.Config) *CommendsService {
	return &CommendsService{cfg: cfg}
}

func (c CommendsService) GetLastPhotoTakenDate() (*time.Time, error) {
	var latestTime time.Time

	files, err := os.ReadDir(c.cfg.OutputDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileInfo, err := os.Stat(filepath.Join(c.cfg.OutputDir, file.Name()))
		if err != nil {
			return nil, err
		}

		fileTime := fileInfo.ModTime()
		if fileTime.After(latestTime) {
			latestTime = fileTime
		}
	}

	return &latestTime, nil
}

func (c CommendsService) RemoveAllPhotos() error {
	// Check if the directory exists
	_, err := os.Stat(c.cfg.OutputDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", c.cfg.OutputDir)
	}

	// Read the directory contents
	files, err := os.ReadDir(c.cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", c.cfg.OutputDir, err)
	}

	type FileData struct {
		Info os.FileInfo
		Path string
	}

	var fileDataList []FileData

	// Filter files older than 10 minutes
	for _, file := range files {
		modInfo, err := file.Info()
		if err != nil {
			log.Printf("Failed to get file info: %s", err)
			continue
		}

		if time.Since(modInfo.ModTime()).Minutes() > 10 {
			fileDataList = append(fileDataList, FileData{Info: modInfo, Path: filepath.Join(c.cfg.OutputDir, file.Name())})
		}
	}

	// Sort files by modification time
	sort.Slice(fileDataList, func(i, j int) bool {
		return fileDataList[i].Info.ModTime().Before(fileDataList[j].Info.ModTime())
	})

	// Delete all but the 10 newest files
	if len(fileDataList) > 10 {
		for _, fileData := range fileDataList[:len(fileDataList)-10] {
			err := os.Remove(fileData.Path)
			if err != nil {
				log.Printf("Failed to remove file: %s", fileData.Path)
			}
		}
	}

	return nil
}
