// Package app is the contract between the Go backend and the web frontend.
// Every exported type here crosses the bridge as JSON, and `mygo generate`
// turns it into a TypeScript interface in src/mygo.ts.
package app

import "time"

// Version is the app's version, shown in the about panel.
const Version = "0.1.0"

// User is the logged-in Bilibili account.
type User struct {
	UID  int64  `json:"uid"`
	Name string `json:"name"`
	// Face is the avatar URL (load it with referrerpolicy="no-referrer").
	Face string `json:"face"`
}

// AuthState tells whether the app holds a valid login.
type AuthState struct {
	LoggedIn bool  `json:"loggedIn"`
	User     *User `json:"user"`
}

// QRStage is a step of a QR code flow (login or face verification).
type QRStage string

const (
	QRWaiting QRStage = "waiting" // shown, not scanned yet
	QRScanned QRStage = "scanned" // scanned, waiting for confirmation on the phone
	QRExpired QRStage = "expired" // the code expired; the flow ends
	QRDone    QRStage = "done"    // confirmed; the flow ends
)

// QRCode is a QR code to show the user.
type QRCode struct {
	// URL is what the code encodes.
	URL string `json:"url"`
	// SVG is the code rendered as a standalone <svg> document, black on transparent.
	SVG string `json:"svg"`
}

// QRUpdate is one step of a QR flow, streamed through a channel.
// The first update of a flow always carries Code.
type QRUpdate struct {
	Stage   QRStage `json:"stage"`
	Code    *QRCode `json:"code"`
	Message string  `json:"message"`
}

// LiveStatus is the state of the live room.
type LiveStatus string

const (
	LiveOffline LiveStatus = "offline" // 未开播
	LiveOn      LiveStatus = "live"    // 直播中
	LiveRound   LiveStatus = "round"   // 轮播中
)

// RoomInfo describes the user's live room.
type RoomInfo struct {
	RoomID         int64      `json:"roomId"`
	UID            int64      `json:"uid"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Cover          string     `json:"cover"`
	AreaID         int64      `json:"areaId"`
	AreaName       string     `json:"areaName"`
	ParentAreaID   int64      `json:"parentAreaId"`
	ParentAreaName string     `json:"parentAreaName"`
	Status         LiveStatus `json:"status"`
	Online         int64      `json:"online"`
	Attention      int64      `json:"attention"`
	// LiveSince is when the current broadcast started, null when offline.
	LiveSince *time.Time `json:"liveSince"`
}

// Area is a sub area (分区) a room can broadcast in.
type Area struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// AreaGroup is a parent area with its sub areas.
type AreaGroup struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Areas []Area `json:"areas"`
}

// StreamInfo is the RTMP address and key to paste into OBS.
type StreamInfo struct {
	Addr string `json:"addr"`
	Code string `json:"code"`
	// Changed is true when Code differs from the previous broadcast's,
	// so the user must update OBS.
	Changed bool `json:"changed"`
}

// RoomSettings are the saved title and area used for the next broadcast.
type RoomSettings struct {
	Title  string `json:"title"`
	AreaID int64  `json:"areaId"`
}

// Preferences are app options the user changes in the settings dialog.
type Preferences struct {
	// Notify shows a desktop notification for SCs and guards.
	Notify bool `json:"notify"`
	// NotifyGiftMin also notifies for gifts worth at least this many gold
	// coins; 0 turns gift notifications off.
	NotifyGiftMin int64 `json:"notifyGiftMin"`
	// Blocked hides danmaku containing any of these words (case-insensitive).
	Blocked []string `json:"blocked"`
}

// StartStage is a step of starting a broadcast.
type StartStage string

const (
	StartRequesting StartStage = "requesting" // calling startLive
	StartFaceAuth   StartStage = "faceAuth"   // face verification needed; Face carries the QR
	StartLive       StartStage = "live"       // broadcast started; Stream is set
)

// StartUpdate is one step of Live.Start, streamed through a channel.
type StartUpdate struct {
	Stage   StartStage  `json:"stage"`
	Face    *QRUpdate   `json:"face"`
	Stream  *StreamInfo `json:"stream"`
	Message string      `json:"message"`
}

// GuardLevel is a captain rank (大航海).
type GuardLevel int

const (
	GuardNone    GuardLevel = 0
	GuardZongdu  GuardLevel = 1 // 总督
	GuardTidu    GuardLevel = 2 // 提督
	GuardJianzha GuardLevel = 3 // 舰长
)

// Medal is a fan medal (粉丝勋章).
type Medal struct {
	Name  string `json:"name"`
	Level int    `json:"level"`
}

// MessageKind is the kind of a danmaku feed item.
type MessageKind string

const (
	KindDanmaku MessageKind = "danmaku" // a viewer's comment
	KindSelf    MessageKind = "self"    // a comment the user sent
	KindGift    MessageKind = "gift"    // a gift or guard purchase
	KindEnter   MessageKind = "enter"   // a viewer entered (欢迎…)
	KindNotice  MessageKind = "notice"  // a room or platform notice
	KindSystem  MessageKind = "system"  // the app's own status lines

	KindSuperChat MessageKind = "superchat" // 醒目留言; SuperChat is set
	KindGuard     MessageKind = "guard"     // 上舰 / 续费; Guard and Gift are set
	KindFollow    MessageKind = "follow"    // a viewer followed the room
	KindLike      MessageKind = "like"      // a viewer liked (点赞)
)

// Message is one item of the danmaku feed.
type Message struct {
	ID         string      `json:"id"`
	Kind       MessageKind `json:"kind"`
	Time       time.Time   `json:"time"`
	UID        int64       `json:"uid"`
	Name       string      `json:"name"`
	Text       string      `json:"text"`
	Admin      bool        `json:"admin"`
	Guard      GuardLevel  `json:"guard"`
	UserLevel  int         `json:"userLevel"`
	Medal      *Medal      `json:"medal"`
	Gift       *Gift       `json:"gift"`
	Historical bool        `json:"historical"` // from the history API, not live

	// Face is the sender's avatar URL, "" when unknown.
	Face string `json:"face"`
	// Emoticon is set when the danmaku is a single sticker (表情包弹幕).
	Emoticon *Emoticon `json:"emoticon"`
	// Emots maps inline emote tokens in Text, such as "[dog]", to images.
	Emots map[string]Emoticon `json:"emots"`
	// SuperChat is set for KindSuperChat.
	SuperChat *SuperChat `json:"superChat"`
}

// Emoticon is an image shown in a danmaku or picked in the emote picker.
type Emoticon struct {
	// Token is what to send to post it: the emoticon_unique for stickers,
	// or the text form such as "[dog]" for inline emotes.
	Token  string `json:"token"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// EmoticonPackage is a group of emoticons available in the room.
type EmoticonPackage struct {
	ID    int64      `json:"id"`
	Name  string     `json:"name"`
	Cover string     `json:"cover"`
	Items []Emoticon `json:"items"`
}

// SuperChat is a 醒目留言.
type SuperChat struct {
	// Price in yuan.
	Price int `json:"price"`
	// Seconds the SC stays pinned.
	Duration int `json:"duration"`
	// Background colour suggested by the server, like "#EDF5FF"; may be "".
	Color string `json:"color"`
}

// Gift is the gift of a KindGift message.
type Gift struct {
	Name string `json:"name"`
	Num  int    `json:"num"`
	// Price is the total value in gold coins (1000 = ¥1); 0 for free gifts.
	Price int64 `json:"price"`
}

// ConnState is the state of the danmaku connection.
type ConnState string

const (
	ConnConnecting   ConnState = "connecting"
	ConnConnected    ConnState = "connected"
	ConnReconnecting ConnState = "reconnecting"
	ConnClosed       ConnState = "closed"
)

// FeedEvent is an item of the Danmaku.Watch channel: a message, or a change
// of connection state.
type FeedEvent struct {
	Message *Message   `json:"message"`
	State   *ConnState `json:"state"`
	// Popularity is the room's popularity from heartbeats, 0 when unknown.
	Popularity int64 `json:"popularity"`
	// Watched is the 看过 count (WATCHED_CHANGE), 0 when not reported.
	Watched int64 `json:"watched"`
	// Online is the high-energy online count (ONLINE_RANK_COUNT), 0 when not reported.
	Online int64 `json:"online"`
	// Likes is the room's total like count (LIKE_INFO_V3_UPDATE), 0 when not reported.
	Likes int64 `json:"likes"`
	// Stats is sent, at most once a second, whenever the session totals change.
	Stats *SessionStats `json:"stats"`
}

// SessionStats are the totals counted since Danmaku.Watch connected
// (live messages only, history excluded).
type SessionStats struct {
	Since      time.Time `json:"since"`
	Danmaku    int       `json:"danmaku"`
	Senders    int       `json:"senders"` // distinct danmaku senders
	Enters     int       `json:"enters"`
	Follows    int       `json:"follows"`
	Likes      int       `json:"likes"`
	Gifts      int       `json:"gifts"`
	GiftValue  int64     `json:"giftValue"` // gold coins of paid gifts, guards and SCs (1000 = ¥1)
	SuperChats int       `json:"superChats"`
	Guards     int       `json:"guards"`
	PeakOnline int64     `json:"peakOnline"`
	// Samples is the online count once a minute, oldest first, at most 240.
	Samples []OnlineSample `json:"samples"`
}

// OnlineSample is the online count at a time.
type OnlineSample struct {
	Time   time.Time `json:"time"`
	Online int64     `json:"online"`
}

// ViewerCard describes a viewer, for the card shown when clicking a name.
type ViewerCard struct {
	UID       int64      `json:"uid"`
	Name      string     `json:"name"`
	Face      string     `json:"face"`
	Sign      string     `json:"sign"`
	Level     int        `json:"level"` // main-site level
	Guard     GuardLevel `json:"guard"` // in this room
	Medal     *Medal     `json:"medal"` // of this room, if worn
	Followers int64      `json:"followers"`
	Muted     bool       `json:"muted"` // currently muted in this room
}

// MutedUser is an entry of the room's mute (禁言) list.
type MutedUser struct {
	UID   int64  `json:"uid"`
	Name  string `json:"name"`
	Admin string `json:"admin"` // who muted them
	// Until is when the mute ends, null when permanent.
	Until *time.Time `json:"until"`
	// ID identifies the entry for Unmute (block_id / tuid as the API needs).
	ID int64 `json:"id"`
}

// RankUser is a row of the contribution (高能榜) or guard (大航海) list.
type RankUser struct {
	Rank  int        `json:"rank"`
	UID   int64      `json:"uid"`
	Name  string     `json:"name"`
	Face  string     `json:"face"`
	Score int64      `json:"score"` // contribution value; 0 for guards
	Guard GuardLevel `json:"guard"`
	Medal *Medal     `json:"medal"`
}

// RankKind picks a list for Live.Rank.
type RankKind string

const (
	RankOnline RankKind = "online" // 高能榜 (current viewers by contribution)
	RankGuard  RankKind = "guard"  // 大航海 members
)

// LogLevel is the severity of a log entry.
type LogLevel string

const (
	LogInfo  LogLevel = "info"
	LogOK    LogLevel = "ok"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

// LogEntry is a line of the activity log.
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   LogLevel  `json:"level"`
	Message string    `json:"message"`
}

// AppInfo describes the build.
type AppInfo struct {
	Version string `json:"version"`
	// Demo is true when the app runs with fake data (BLH_DEMO=1).
	Demo bool `json:"demo"`
	// ConfigPath is where the login and settings are stored.
	ConfigPath string `json:"configPath"`
}
