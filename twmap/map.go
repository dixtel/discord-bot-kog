package twmap

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var VALID_MAP_FILE_NAME = regexp.MustCompile(`^[a-zA-Z_\-0-9]+\.map$`)

func IsValidMapFileName(filename string) bool {
	return VALID_MAP_FILE_NAME.Match([]byte(filename))
}

func RemoveMapFileExtension(filename string) string {
	if filename[len(filename)-4:] == ".map" {
		return filename[:len(filename)-4]
	}

	log.Warn().Str("filename", filename).Msg("this should not happen")

	return filename
}

func DownloadMapFromDiscord(mapUrl string) ([]byte, error) {
	res, err := http.Get(mapUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot http get: %w", err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read all: %w", err)
	}

	return b, nil
}

func MakeScreenshot(mapSource []byte) ([]byte, error) {
	dir := fmt.Sprintf("/tmp/%s", uuid.New().String())

	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cannot crate temp directory: %w", err)
	}

	err = saveFile(mapSource, dir+"/input.map")
	if err != nil {
		return nil, fmt.Errorf("cannot download map %w", err)
	}

	currDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("cannot get working directory %w", err)
	}

	args := []string{"twgpu-map-photography", "input.map", currDir + "/res/test-map/mapres"}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("cannot execute command %v: %w: %s", args, err, output)
	}

	file, err := os.ReadFile(dir + "/input.png")
	if err != nil {
		return nil, fmt.Errorf("cannot read screenshot file %w", err)
	}

	return file, nil
}

func saveFile(source []byte, path string) error {
	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}

	defer out.Close()

	_, err = io.Copy(out, bytes.NewReader(source))
	if err != nil {
		return fmt.Errorf("cannot copy: %w", err)
	}

	return nil
}
