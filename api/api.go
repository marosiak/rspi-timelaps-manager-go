package api

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/macrosiak/rspi-timelaps-manager-go/config"
	. "github.com/macrosiak/rspi-timelaps-manager-go/system_stats"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Api struct {
	cfg            *config.Config
	systemStatsSrv *SystemStatsService
}

func NewApi(app *fiber.App, systemStatsSrv *SystemStatsService) *Api {
	api := &Api{cfg: config.New(), systemStatsSrv: systemStatsSrv}

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Static("/", api.cfg.WebInterfaceFilesPath)

	app.Get("/ws/", websocket.New(api.WebsocketHandler))
	return api
}

type WebsocketResponse struct {
	Stats            *StatsResponse `json:"stats"`
	LastPhotoTakenAt *int64         `json:"lastPhotoTakenAt"`
}

func (a Api) getLastPhotoTakenAt() (*time.Time, error) {
	var latestTime time.Time

	files, err := os.ReadDir(a.cfg.OutputDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileInfo, err := os.Stat(filepath.Join(a.cfg.OutputDir, file.Name()))
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

func (a Api) removeAllPhotos() error {
	// Check if the directory exists
	_, err := os.Stat(a.cfg.OutputDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", a.cfg.OutputDir)
	}

	// Read the directory contents
	files, err := os.ReadDir(a.cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", a.cfg.OutputDir, err)
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
			fileDataList = append(fileDataList, FileData{Info: modInfo, Path: filepath.Join(a.cfg.OutputDir, file.Name())})
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

	fmt.Println("Operation completed successfully!")
	return nil
}

const (
	ActionRemoveAllImages = "REMOVE_ALL_IMAGES"
)

type Action string
type ActionRequest struct {
	Action Action `json:"action"`
}

func (a Api) WebsocketHandler(c *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Err(err).Msg("read message")
			break
		}

		if len(msg) > 0 {
			actionRequest := ActionRequest{}
			err := json.Unmarshal(msg, &actionRequest)
			if err != nil {
				log.Err(err).Str("msg", string(msg)).Msg("unmarshal")
			}
			if actionRequest.Action == ActionRemoveAllImages {
				err := a.removeAllPhotos()
				if err != nil {
					log.Err(err).Msg("remove all images")
				}
			}
		}

		systemInfo, err := a.systemStatsSrv.GetSystemUsageInfo()
		if err != nil {
			log.Err(err).Msg("get system info")
			continue
		}

		lastPhotoTakenAt, err := a.getLastPhotoTakenAt()
		if err != nil {
			log.Err(err).Msg("get last photo taken at")
			continue
		}

		lastPhotoTakenAtTimestamp := lastPhotoTakenAt.Unix()
		response := WebsocketResponse{
			Stats:            systemInfo,
			LastPhotoTakenAt: &lastPhotoTakenAtTimestamp,
		}

		respJson, err := json.Marshal(response)
		if err != nil {
			log.Err(err).Msg("json marshal")
			break
		}

		if err = c.WriteMessage(mt, respJson); err != nil {
			log.Err(err).Msg("write message")
			break
		}
	}
}
