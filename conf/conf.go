package conf

var BotToken string
var ChannelName string
var Pass string
var Mode string
var BaseUrl string

type UploadResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const FileRoute = "/api/download/"
