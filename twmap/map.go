package twmap

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"

	"github.com/dixtel/dicord-bot-kog/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var FIXED_MAP_NAME_REGEX = regexp.MustCompile(`[a-zA-Z_\-0-9]`)

func FixMapName(name string) string {
	fixed := ""
	for _, ch := range name {
		if FIXED_MAP_NAME_REGEX.Match([]byte{byte(ch)}) {
			fixed += string(ch)
		}
	}
	return fixed
}

//TODO: test this function
func MapNameExists(db *gorm.DB, fixedMapName string) (bool, error) {
	record := &models.Map{
		FixedName: fixedMapName,
	}
	res := db.First(record)
	if res.Error != nil {
		return false, fmt.Errorf("cannot get first record: %w", res.Error)
	}

	return res.RowsAffected > 0, nil
}

func MakeScreenshot(mapUrl string) (io.ReadCloser, error) {
	dir := fmt.Sprintf("/tmp/%s", uuid.New().String())
	
	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("cannot crate temp directory: %w", err)
	}

	err = downloadMap(mapUrl, dir + "/input.map")
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

	_, err = cmd.Output()
	if err != nil { 
		return nil, fmt.Errorf("cannot execute command %v: %w", args, err)
	}

	file ,err := os.Open(dir + "/input.png")
	if err != nil { 
		return nil, fmt.Errorf("cannot read screenshot file %w", err)
	}

	return file, nil
}

func downloadMap(url string, outputPath string) error {
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot http get: %w", err)
	}
	defer res.Body.Close()

	out, err := os.Create(outputPath)
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return fmt.Errorf("cannot copy: %w", err)
	}

	return nil
}