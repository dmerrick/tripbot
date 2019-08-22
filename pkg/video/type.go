package video

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dmerrick/danalol-stream/pkg/config"
	"github.com/dmerrick/danalol-stream/pkg/database"
	"github.com/dmerrick/danalol-stream/pkg/ocr"
)

// Videos represent a video file containing dashcam footage
type Video struct {
	Id          int           `db:"id"`
	Slug        string        `db:"slug"`
	Lat         float64       `db:"lat"`
	Lng         float64       `db:"lng"`
	NextVid     sql.NullInt64 `db:"next_vid"`
	PrevVid     sql.NullInt64 `db:"prev_vid"`
	Flagged     bool          `db:"flagged"`
	DateFilmed  time.Time     `db:"date_filmed"`
	DateCreated time.Time     `db:"date_created"`
}

// LoadOrCreate() will look up the video in the DB,
// or add it to the DB if it's not there yet
func LoadOrCreate(path string) (Video, error) {
	slug := slug(path)

	vid, err := load(slug)
	if err != nil {
		// create a new video
		vid, err = create(slug)
	}

	return vid, err
}

// Location returns a lat/lng pair
//TODO: refactor out the error return value
func (v Video) Location() (float64, float64, error) {
	var err error
	if v.Flagged {
		err = errors.New("video is flagged")
	}
	return v.Lat, v.Lng, err
}

// ex: 2018_0514_224801_013_a_opt
func (v Video) String() string {
	return v.Slug
}

// a DashStr is the string we get from the dashcam
// an example file: 2018_0514_224801_013.MP4
// an example dashstr: 2018_0514_224801_013
// ex: 2018_0514_224801_013
func (v Video) DashStr() string {
	//TODO: this never should have happened, but it did and it crashed the bot
	if len(v.Slug) < 20 {
		return ""
	}
	return v.Slug[:20]
}

// ex: 2018_0514_224801_013.MP4
func (v Video) File() string {
	return fmt.Sprintf("%s.MP4", v.Slug)
}

// ex: /Volumes/.../2018_0514_224801_013.MP4
func (v Video) Path() string {
	return path.Join(config.VideoDir(), v.File())
}

// toDate parses the vidStr and returns a time.Time object for the video
func (v Video) toDate() time.Time {
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

// ocrCoords will use OCR to read the coordinates from a screenshot (seriously)
func (v Video) ocrCoords() (float64, float64, error) {
	for _, timestamp := range timestampsToTry {
		lat, lon, err := ocr.CoordsFromImage(v.screencap(timestamp))
		if err == nil {
			return lat, lon, err
		}
	}
	return 0, 0, errors.New("none of the screencaps had valid coords")
}

func load(slug string) (Video, error) {
	var newVid Video
	// try to find the slug in the DB
	videos := []Video{}
	query := fmt.Sprintf("SELECT * FROM videos WHERE slug='%s'", slug)
	err := database.DBCon.Select(&videos, query)
	if err != nil {
		log.Println("error fetching vid from DB:", err)
		return newVid, err
	}

	// did we find anything in the DB?
	if len(videos) == 0 {
		err = errors.New("no matches found")
		return newVid, err
	}
	return videos[0], nil
}

// create will create a new Video from a slug
//TODO: this is kinda weird, we create an empty Video
// and then we save it to the DB... maybe we could just
// save right to the DB? It would take some refactoring.
func create(file string) (Video, error) {
	var newVid Video
	var blankDate time.Time

	if file == "" {
		return newVid, errors.New("no file provided")
	}
	slug := slug(file)

	// validate the dash string
	err := validate(slug)
	if err != nil {
		return newVid, err
	}

	// create new (mostly) empty vid
	newVid = Video{
		Slug:        slug,
		Lat:         0,
		Lng:         0,
		Flagged:     false,
		DateFilmed:  blankDate,
		DateCreated: blankDate,
	}

	// store the video in the DB
	err = newVid.save()
	if err != nil {
		log.Println("error saving to DB:", err)
	}

	// now fetch it from the DB
	//TODO: this is an extra DB call, do we care?
	dbVid, err := load(slug)

	return dbVid, err
}

// save() will store the video in the DB
func (v Video) save() error {
	flagged := false
	// try to get at least one good coords pair
	lat, lng, err := v.ocrCoords()
	if err != nil {
		log.Println("error fetching coords:", err)
		flagged = true
	}

	tx := database.DBCon.MustBegin()
	tx.MustExec(
		"INSERT INTO videos (slug, lat, lng, date_filmed, flagged) VALUES ($1, $2, $3, $4, $5)",
		v.Slug,
		lat,
		lng,
		v.toDate(),
		flagged,
	)
	return tx.Commit()
}

// slug strips the path and extension off the file
func slug(file string) string {
	fileName := path.Base(file)
	return removeFileExtension(fileName)
}

// these are different timestamps we have screenshots prepared for
// the "000" corresponds to 0m0s, "130" corresponds to 1m30s
var timestampsToTry = []string{
	"000",
	"015",
	"030",
	"045",
	"100",
	"115",
	"130",
	"145",
	"200",
	"215",
	"230",
	"245",
}

// timestamp is something like 000, 030, 100, etc
func (v Video) screencap(timestamp string) string {
	screencapFile := fmt.Sprintf("%s-%s.png", v.DashStr(), timestamp)
	return path.Join(config.ScreencapDir(), timestamp, screencapFile)
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