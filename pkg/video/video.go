package video

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dmerrick/danalol-stream/pkg/config"
	"github.com/dmerrick/danalol-stream/pkg/helpers"
)

// a DashStr is the string we get from the dashcam
// an example file: 2018_0514_224801_013.MP4
// an example dashstr: 2018_0514_224801_013
type Video struct {
	DashStr string
}

func New(file string) (Video, error) {
	var newVid Video

	if file == "" {
		return newVid, errors.New("no file provided")
	}
	fileName := path.Base(file)
	dashStr := removeFileExtension(fileName)

	// validate the dash string
	err := validate(dashStr)
	if err != nil {
		return newVid, err
	}
	newVid = Video{dashStr}
	return newVid, err
}

// ex: 2018_0514_224801_013
func (v Video) String() string {
	return v.DashStr
}

// ex: 2018_0514_224801_013.MP4
func (v Video) File() string {
	return fmt.Sprintf("%s.MP4", v.DashStr)
}

// ex: /Volumes/.../2018_0514_224801_013.MP4
func (v Video) Path() string {
	return path.Join(config.VideoDir, v.File())
}

func (v Video) Date() time.Time {
	vidStr := v.String()
	year, _ := strconv.Atoi(vidStr[:4])
	month, _ := strconv.Atoi(vidStr[5:7])
	day, _ := strconv.Atoi(vidStr[7:9])
	hour, _ := strconv.Atoi(vidStr[10:12])
	minute, _ := strconv.Atoi(vidStr[12:14])
	second, _ := strconv.Atoi(vidStr[14:16])

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return t
}

func CurrentlyPlaying() string {
	// run the shell script to get currently-playing video
	scriptPath := path.Join(helpers.ProjectRoot(), "bin/current-file.sh")
	out, err := exec.Command(scriptPath).Output()
	if err != nil {
		log.Printf("failed to run script: %v", err)
	}
	return string(out)
}

func validate(dashStr string) error {
	if len(dashStr) < 20 {
		return errors.New("dash string too short")
	}
	shortened := dashStr[:20]

	if strings.HasPrefix(".", shortened) {
		return errors.New("dash string can't be a hidden file")
	}

	//TODO: this should probably live in an init()
	var validDashStr = regexp.MustCompile(`^[_0-9]{20}$`)
	if !validDashStr.MatchString(shortened) {
		return errors.New("dash string did not match regex")
	}
	return nil
}

func removeFileExtension(filename string) string {
	ext := path.Ext(filename)
	return filename[0 : len(filename)-len(ext)]
}
