package video

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/adanalife/tripbot/pkg/database"
	terrors "github.com/adanalife/tripbot/pkg/errors"
	"github.com/adanalife/tripbot/pkg/helpers"
)

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

// load() fetches a Video from the DB
func load(slug string) (Video, error) {
	//TODO: consider replacing this with a &Video{},
	// perhaps Video{ID:0}?
	var newVid Video
	// try to find the slug in the DB
	videos := []Video{}
	query := `SELECT * FROM videos WHERE slug=$1`
	err := database.Connection().Select(&videos, query, slug)
	if err != nil {
		terrors.Log(err, "error fetching vid from DB")
		return newVid, err
	}

	// did we find anything in the DB?
	if len(videos) == 0 {
		err = errors.New("no matches found")
		return newVid, err
	}
	return videos[0], nil
}

//TODO: combine this with load()?
func loadById(id int64) (Video, error) {
	var newVid Video
	// try to find the slug in the DB
	videos := []Video{}
	query := `SELECT * FROM videos WHERE id=$1`
	err := database.Connection().Select(&videos, query, id)
	if err != nil {
		terrors.Log(err, "error fetching vid from DB")
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
		terrors.Log(err, "error saving to DB")
	}

	// now fetch it from the DB
	//TODO: this is an extra DB call, do we care?
	dbVid, err := load(slug)

	return dbVid, err
}

// save() will store the video in the DB
//TODO: I think this can be achieved much easier, c.p. user save
func (v Video) save() error {
	var err error
	flagged := v.Flagged
	lat := v.Lat
	lng := v.Lng
	state := v.State

	if lat == 0 || lng == 0 {
		//TODO: this is where we used to run ocrCoords()
		terrors.Log(nil, "OCRing coords skipped!")
		flagged = true
	}

	if !flagged {
		// figure out which state we're in
		state, err = helpers.StateFromCoords(lat, lng)
		if err != nil {
			terrors.Log(err, "error geocoding coords")
		}
	}

	tx := database.Connection().MustBegin()
	//TODO: do something with result var here?
	_, err = tx.Exec(
		"INSERT INTO videos (slug, lat, lng, date_filmed, flagged, prev_vid, next_vid, state) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		v.Slug,
		lat,
		lng,
		v.toDate(),
		flagged,
		v.PrevVid,
		v.NextVid,
		state,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Next() finds the next unflagged video
//TODO: should this be NextUnflagged?
//TODO: handle errors in here?
func (v Video) Next() Video {
	vid := v
	for { // ever
		vid, _ = loadById(vid.NextVid.Int64)
		// use the first unflagged video we find
		if !vid.Flagged {
			break
		}
	}
	return vid
}

func (v Video) SetNextVid(nextVid Video) error {
	_, err := database.Connection().NamedExec(`UPDATE videos SET next_vid=:next WHERE id = :id`,
		map[string]interface{}{
			"next": nextVid.Id,
			"id":   v.Id,
		})
	return err
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

func FindRandomByState(state string) (Video, error) {
	var newVid Video

	// convert to long form
	if len(state) == 2 {
		state = helpers.StateAbbrevToState(state)
		if state == "" {
			return newVid, fmt.Errorf("unable to parse state abbrev")
		}
	}
	// title-case the state (it's stored in the DB like that)
	state = helpers.TitlecaseState(state)

	// try to find the slug in the DB
	videos := []Video{}
	//TODO: ORDER BY random() will eventually get too slow
	query := `SELECT * FROM videos WHERE state=$1 ORDER BY random() LIMIT 1`
	err := database.Connection().Select(&videos, query, state)
	if err != nil {
		terrors.Log(err, "error fetching vid from DB")
		return newVid, err
	}

	// did we find anything in the DB?
	if len(videos) == 0 {
		return newVid, &terrors.NoFootageForStateError{Msg: "no matches found"}
	}
	return videos[0], nil
}
