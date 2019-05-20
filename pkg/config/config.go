package config

const (
	// BotUsername is the name of the bot
	BotUsername = "TripBot4000"
	// ChannelName is the channel to join
	ChannelName = "adanalife_"

	// GetCurrentVidScript is a script that figures out the currently-playing video
	//TODO this should use the ProjectRoot helper
	GetCurrentVidScript = "/Users/dmerrick/other_projects/danalol-stream/bin/current-file.sh"

	//TODO move these off usbshare1
	ScreencapDir  = "/Volumes/usbshare1/screencaps"
	VideoDir      = "/Volumes/usbshare1/Dashcam/_all"
	MapsOutputDir = "/Volumes/usbshare1/maps"
	CroppedPath   = "/Volumes/usbshare1/cropped-corners"

	//TODO capitalize me
	DbPath            = "tripbot.db"
	UserJoinsBucket   = "user_joins"
	UserWatchedBucket = "user_watched"
	CoordsBucket      = "coords"
)

//TODO: this should load from a config file
// IgnoredUsers are users who shouldn't be in the running for miles
// https://twitchinsights.net/bots
var IgnoredUsers = []string{
	"adanalife_",
	"tripbot4000",
	"nightbot",
	"anotherttvviewer",
	"apricotdrupefruit",
	"avocadobadado",
	"commanderroot",
	"communityshowcase",
	"electricallongboard",
	"eubyt",
	"feuerwehr",
	"freddyybot",
	"jobi_essen",
	"logviewer",
	"lurxx",
	"p0lizei_",
	"slocool",
	"taormina2600",
	"unixchat",
	"v_and_k",
	"virgoproz",
	"zanekyber",
}

// HelpMessages are all of the different things !help can return
var HelpMessages = []string{
	"!tripbot: Get the current location (beta)",
	"!map: Show a map of the whole trip",
	"!info: Get more details on the footage",
	"!song: Get the current music",
	"!miles: See your current miles",
}

var GoogleMapsStyle = []string{
	"element:geometry|color:0x242f3e",
	"element:labels.text.stroke|lightness:-80",
	"feature:administrative|element:labels.text.fill|color:0x746855",
	"feature:administrative.locality|element:labels.text.fill|color:0xd59563",
	"feature:poi|element:labels.text.fill|color:0xd59563",
	"feature:poi.park|element:geometry|color:0x263c3f",
	"feature:poi.park|element:labels.text.fill|color:0x6b9a76",
	"feature:road|element:geometry.fill|color:0x2b3544",
	"feature:road|element:labels.text.fill|color:0x9ca5b3",
	"feature:road.arterial|element:geometry.fill|color:0x38414e",
	"feature:road.arterial|element:geometry.stroke|color:0x212a37",
	"feature:road.arterial|element:labels|visibility:off",
	"feature:road.highway|element:geometry.fill|color:0x746855",
	"feature:road.highway|element:geometry.stroke|color:0x1f2835",
	"feature:road.highway|element:labels|visibility:off",
	"feature:road.highway|element:labels.text.fill|color:0xf3d19c",
	"feature:road.local|visibility:off",
	"feature:road.local|element:geometry.fill|color:0x38414e",
	"feature:road.local|element:geometry.stroke|color:0x212a37",
	"feature:transit|element:geometry|color:0x2f3948",
	"feature:transit.station|element:labels.text.fill|color:0xd59563",
	"feature:water|element:geometry|color:0x17263c",
	"feature:water|element:labels.text.fill|color:0x515c6d",
	"feature:water|element:labels.text.stroke|lightness:-20",
}
