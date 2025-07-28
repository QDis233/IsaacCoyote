package isaac

import (
	"IsaacCoyote/util"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func waitForProcess(processName string) uint32 {
	for {
		pid, err := util.GetProcPID(processName)
		if err != nil {
			zap.L().Error(fmt.Sprintf("未找到 [%s] 进程", processName))
			time.Sleep(5 * time.Second)
			continue
		}
		return pid
	}
}

func getModDataFile(isaacPID uint32) (string, error) {
	procFilePath, err := util.GetProcPath(isaacPID)
	if err != nil {
		return "", err
	}
	isaacDir := filepath.Dir(procFilePath)
	modDataPath := filepath.Join(isaacDir, "data", "isaac-coyote")

	var latestModTime int64
	var latestDataFile string
	entries, err := os.ReadDir(modDataPath)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "save") && strings.HasSuffix(entry.Name(), ".dat") {
			filePath := path.Join(modDataPath, entry.Name())
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return "", err
			}

			if fileInfo.ModTime().Unix() > latestModTime {
				latestModTime = fileInfo.ModTime().Unix()
				latestDataFile = filePath
			}
		}
	}
	return latestDataFile, nil
}
