package api

import (
	"net/http"
	"os"
	"strings"

	"lanyundev/tgstate/conf"
	"lanyundev/tgstate/control"
)

func Vercel(w http.ResponseWriter, r *http.Request) {
	conf.BotToken = os.Getenv("token")
	conf.ChannelName = os.Getenv("target")
	conf.Pass = os.Getenv("pass")
	conf.Mode = os.Getenv("mode")
	// 获取请求路径
	path := r.URL.Path
	// 如果请求路径以 "/img/" 开头
	if strings.HasPrefix(path, conf.FileRoute) {
		control.DownloadAPI(w, r)
		return // 结束处理，确保不执行默认处理
	}
	switch path {
	case "/api":
		// 调用 control 包中的 UploadAPI 处理函数
		control.Middleware(control.UploadAPI)(w, r)
	case "/pwd":
		control.Pwd(w, r)
	default:
		control.Middleware(control.Index)(w, r)
	}
}
