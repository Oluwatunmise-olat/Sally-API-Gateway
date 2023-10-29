package files

import (
	"fmt"
	"github.com/Oluwatunmise-olat/custom-api-gateway/pkg/logger"
	"io"
	"mime/multipart"
	"os"
	"strings"
)

func CreateDirectoryIfNotExists(directoryName string) {
	if _, err := os.Stat(directoryName); os.IsNotExist(err) {
		os.Mkdir(directoryName, 0755)
	}
}

func CheckIfPathExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// Check if route is in form of a regex pattern identified by the system i.e {route+} and transforms such route
// ie. /github/{account+} would be transformed into /github/account/*
func CheckIfIsRegexPathAndTransformRoute(path string) (bool, string) {
	configRegexChars := []string{"{", "+", "}"}

	listeningPathSlice := strings.Split(path, "/")
	trailingPath := listeningPathSlice[len(listeningPathSlice)-1]

	if strings.Contains(trailingPath, "+") {
		for _, configChar := range configRegexChars {
			trailingPath = strings.Replace(trailingPath, configChar, "", -1)
		}

		trailingPath = strings.Join(listeningPathSlice[:len(listeningPathSlice)-1], "/") + fmt.Sprintf("/%s/{route:.*}", trailingPath)

		return true, trailingPath
	} else {
		return false, path
	}
}

func CreateOrUpdateConfigFile(file multipart.File) error {
	CreateDirectoryIfNotExists("config")

	newConfigFile, err := os.OpenFile("config/app.yaml", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logger.Log.Errorln("Error occurred opening config file directory")
		return err
	}
	defer newConfigFile.Close()

	_, err = io.Copy(newConfigFile, file)
	if err != nil {
		logger.Log.Errorln("Error occurred copying uploaded config file")
		return err
	}

	return nil
}
