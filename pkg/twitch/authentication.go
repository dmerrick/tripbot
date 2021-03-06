package twitch

import (
	"log"
	"os"

	c "github.com/adanalife/tripbot/pkg/config/tripbot"
	terrors "github.com/adanalife/tripbot/pkg/errors"
	"github.com/adanalife/tripbot/pkg/helpers"
	"github.com/logrusorgru/aurora"
	"github.com/nicklaw5/helix"
)

// currentTwitchClient is the standard twitch client
var currentTwitchClient *helix.Client

// these are used to authenticate to twitch
var ClientID string
var ClientSecret string
var AuthToken string
var AppAccessToken string

// these are used to authenticate requests that require user permissions
var UserAccessToken string
var UserRefreshToken string

// init makes sure we have all of the require ENV vars
func init() {
	requiredVars := []string{
		"TWITCH_AUTH_TOKEN",
		"TWITCH_CLIENT_ID",
		"TWITCH_CLIENT_SECRET",
	}
	for _, v := range requiredVars {
		_, ok := os.LookupEnv(v)
		if !ok {
			log.Fatalf("You must set %s", v)
		}
	}
	AuthToken = os.Getenv("TWITCH_AUTH_TOKEN")
	ClientID = os.Getenv("TWITCH_CLIENT_ID")
	ClientSecret = os.Getenv("TWITCH_CLIENT_SECRET")
}

// Client creates a twitch client, or returns the existing one
func Client() (*helix.Client, error) {
	// use the existing client if we have one
	if currentTwitchClient != nil {
		return currentTwitchClient, nil
	}
	client, err := helix.NewClient(&helix.Options{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		// this is set at https://dev.twitch.tv/console/apps
		RedirectURI: c.Conf.ExternalURL + "/auth/callback",
	})
	if err != nil {
		terrors.Log(err, "error creating client")
	}

	//TODO: move to configs lib
	//TODO: revisit that we need all of these
	scopes := []string{"openid", "user:edit:broadcast", "channel:read:subscriptions"}
	// set the AppAccessToken
	resp, err := client.RequestAppAccessToken(scopes)
	if err != nil {
		terrors.Log(err, "error getting app access token from twitch")
	}
	AppAccessToken = resp.Data.AccessToken
	client.SetAppAccessToken(AppAccessToken)

	// use this as the shared client
	currentTwitchClient = client

	return client, err
}

// GenerateUserAccessToken sends a code to Twitch to generate a
// user access token. This is called by the web server after
// going through the OAuth flow
func GenerateUserAccessToken(code string) {
	resp, err := currentTwitchClient.RequestUserAccessToken(code)
	if err != nil {
		terrors.Log(err, "error getting user access token from twitch")
		return
	}

	UserAccessToken = resp.Data.AccessToken
	UserRefreshToken = resp.Data.RefreshToken

	// update the current client with the access token
	currentTwitchClient.SetUserAccessToken(UserAccessToken)
}

// RefreshUserAccessToken makes a call to Twitch to generate a
// fresh user access token. It requires a UserRefreshToken to be
// set already.
func RefreshUserAccessToken() {
	// check to see if we have the required tokens to work with
	if UserRefreshToken == "" || UserAccessToken == "" {
		log.Println("no user access token was present, did you log in with OAuth?")
		authURL := currentTwitchClient.GetAuthorizationURL(&helix.AuthorizationURLParams{
			//TODO: move to configs lib
			Scopes:       []string{"openid", "user:edit:broadcast", "channel:read:subscriptions"},
			ResponseType: "code",
		})
		log.Println(aurora.Blue(authURL).Underline())
		// send a text message cause some features won't work
		// without a user access token
		helpers.SendSMS("refreshing user access token failed!")
		return
	}

	resp, err := currentTwitchClient.RefreshUserAccessToken(UserRefreshToken)
	if err != nil {
		terrors.Log(err, "error refreshing user access token")
		return
	}

	UserAccessToken = resp.Data.AccessToken
	UserRefreshToken = resp.Data.RefreshToken

	// update the current client with the new access token
	currentTwitchClient.SetUserAccessToken(UserAccessToken)

	log.Println(aurora.Cyan("successfully updated user access token"))
}
