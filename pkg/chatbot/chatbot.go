package chatbot

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dmerrick/danalol-stream/pkg/config"
	terrors "github.com/dmerrick/danalol-stream/pkg/errors"
	"github.com/dmerrick/danalol-stream/pkg/helpers"
	mylog "github.com/dmerrick/danalol-stream/pkg/log"
	mytwitch "github.com/dmerrick/danalol-stream/pkg/twitch"
	"github.com/dmerrick/danalol-stream/pkg/users"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/kelvins/geocoder"
	"github.com/logrusorgru/aurora"
	"github.com/nicklaw5/helix"
)

var googleMapsAPIKey string
var client *twitch.Client
var Uptime time.Time

// used to determine which help message to display
// randomized so it starts with a new one every restart
var helpIndex = rand.Intn(len(config.HelpMessages))

const followerMsg = "Follow the stream to run unlimited commands :)"
const subscriberMsg = "You must be a subscriber to run that command :)"

func Initialize() *twitch.Client {
	var err error
	Uptime = time.Now()

	//TODO: remove this, doesn't seem needed
	// load ENV vars from .env file
	// err = godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	// set up geocoder (for translating coords to places)
	geocoder.ApiKey = config.GoogleMapsAPIKey

	// initialize the twitch API client
	c, err := mytwitch.Client()
	if err != nil {
		terrors.Fatal(err, "unable to create twitch API client")
	}

	//TODO: actually use these security features
	authURL := c.GetAuthorizationURL("", false)
	log.Println("if your browser doesn't open automatically:")
	log.Println(aurora.Blue(authURL).Underline())
	helpers.OpenInBrowser(authURL)

	client = twitch.NewClient(config.BotUsername, mytwitch.AuthToken)

	// attach handlers
	client.OnUserJoinMessage(UserJoin)
	client.OnUserPartMessage(UserPart)
	// client.OnUserNoticeMessage(chatbot.UserNotice)
	client.OnWhisperMessage(Whisper)
	client.OnPrivateMessage(PrivateMessage)

	return client
}

// Say will make a post in chat
func Say(msg string) {
	mylog.ChatMsg(config.BotUsername, msg)
	client.Say(config.ChannelName, msg)
}

// Chatter is designed to most a randomized message on a timer
// right now it just posts random "help messages"
func Chatter() {
	// rand.Intn(len(config.HelpMessages))
	Say(help())
}

func help() string {
	text := config.HelpMessages[helpIndex]
	// bump the index
	helpIndex = (helpIndex + 1) % len(config.HelpMessages)
	return text
}

func AnnounceNewFollower(username string) {
	msg := fmt.Sprintf("Thank you for the follow, @%s", username)
	Say(msg)
}

// AnnounceSubscriber makes a post in chat that a user has subscribed
func AnnounceSubscriber(sub helix.Subscription) {
	//TODO: do more with the Subscription... IsGift, Tier, PlanName, etc.
	spew.Dump(sub)
	username := sub.UserName
	msg := fmt.Sprintf("Thank you for the sub, @%s; enjoy your !bonusmiles bleedPurple", username)
	Say(msg)
	// give everyone a bonus mile
	users.GiveEveryoneMiles(1.0)
	msg = fmt.Sprintf("The %d current viewers have been given a bonus mile, too HolidayPresent", len(users.LoggedIn))
	Say(msg)
}
