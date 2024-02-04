package control

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	// "path/filepath"
	"strings"
	// "time"
	"strconv"

	"lanyundev/tgstate/assets"
	"lanyundev/tgstate/conf"
	"lanyundev/tgstate/utils"
)

// 全局常量
const TemplatesPath string = "templates_min"

// UploadAPI 上传文件API处理函数
func UploadAPI(resp http.ResponseWriter, req *http.Request) {
	// 检查请求方法是否为POST
	if req.Method != http.MethodPost {
		// 不是POST请求，返回错误响应
		http.Error(resp, "无效的请求方法!", http.StatusMethodNotAllowed)
		return
	}

	// 设置跨域访问允许的Origin
	resp.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取上传的文件
	uploadedFile, fileHeader, err := req.FormFile("file")
	if err != nil {
		// 处理无法获取文件的错误，返回JSON格式的错误消息
		errJsonMsg("无法获取文件!", resp)
		return
	}
	defer uploadedFile.Close()

	// 在非网盘模式下,检查文件大小是否超过限制（20MB）
	// if conf.Mode != "p" && req.ContentLength > (20*1024*1024+255) {
	// 	errJsonMsg("文件大小超过20MB限制!", resp)
	// 	return
	// }
	if req.ContentLength > (20*1024*1024+255) {
		errJsonMsg("文件大小超过20MB限制!", resp)
		return
	}

	// 上传文件类型检查（注释掉的部分）需要添加: "path/filepath"
	// allowedExts := []string{".jpg", ".jpeg", ".png", ".webp"}
	// ext := filepath.Ext(fileHeader.Filename)
	// valid := false
	// for _, allowedExt := range allowedExts {
	// 	if ext == allowedExt {
	// 		valid = true
	// 		break
	// 	}
	// }
	// if conf.Mode != "p" && !valid {
	// 	errJsonMsg("无效的文件类型。仅允许 .jpg、.jpeg 和 .png 类型的文件。", resp)
	// 	return
	// }

    // 处理文件上传，并获取上传后的文件路径
    uploadedFilePath := conf.FileRoute + utils.UpDocument(utils.TgFileData(fileHeader.Filename, uploadedFile))

    // 构建响应消息
    response := conf.UploadResponse{
        Code:    1,
        Message: uploadedFilePath,
    }
    // 如果文件路径等于默认路径，则上传失败，更新响应消息
    if uploadedFilePath == conf.FileRoute {
        response = conf.UploadResponse{
            Code:    0,
            Message: "后端文件上传失败!",
        }
    }

    // 设置响应头为JSON格式，返回成功状态码并将响应消息编码为JSON格式发送给客户端
    resp.Header().Set("Content-Type", "application/json")
    resp.WriteHeader(http.StatusOK) // 返回 200 响应
    json.NewEncoder(resp).Encode(response)
    
	return
}

// errJsonMsg 生成JSON格式的错误消息并发送给客户端
func errJsonMsg(message string, resp http.ResponseWriter) {
	response := conf.UploadResponse{
		Code:    -1,
		Message: message,
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(resp).Encode(response)
}

// DownloadAPI 处理文件下载请求的函数
func DownloadAPI(w http.ResponseWriter, r *http.Request) {
	// 检查请求方法是否为GET
	if r.Method != http.MethodGet {
		// 不是GET请求，返回错误响应
		http.Error(w, "不允许的请求方法!", http.StatusMethodNotAllowed)
		return
	}

	// 获取请求路径和文件名参数
	path := r.URL.Path
	fileName := r.FormValue("filename")

	// 从路径中提取文件ID
	id := strings.TrimPrefix(path, conf.FileRoute)
	if id == "" {
		// 如果文件ID为空，返回404 Not Found错误
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
		return
	}

	// 发起HTTP GET请求获取Telegram文件
	resp, err := http.Get(utils.GetDownloadUrl(id))
	if err != nil {
		// 处理获取内容失败的错误，返回500 Internal Server Error
		http.Error(w, "获取内容失败", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 获取Content-Length头字段
	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		log.Println("获取Content-Length出错:", err)
		return
	}

	// 根据Content-Length设置缓冲区大小
	buffer := make([]byte, contentLength)
	n, err := resp.Body.Read(buffer)
	if err != nil && err != io.ErrUnexpectedEOF {
		log.Println("读取响应主体数据时发生错误:", err)
		return
	}

	// 判断文件类型是否为tgstate-blob分片类型
	if string(buffer[:12]) == "tgstate-blob" {
		// 解析分片信息
		content := string(buffer)
		lines := strings.Fields(content)
		if len(lines) < 2 {
			log.Println("分片信息格式错误")
			return
		}

		// 获取分片文件名
		fileName = strings.TrimSpace(lines[1])
		log.Println("分块文件:", fileName)

		// 设置响应头，准备下载分片文件
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")

		// 遍历分片信息，获取每个分片文件并写入响应主体
		for i := 2; i < len(lines); i++ {
			blobResp, err := http.Get(utils.GetDownloadUrl(strings.ReplaceAll(lines[i], " ", "")))
			if err != nil {
				http.Error(w, "获取内容失败", http.StatusInternalServerError)
				return
			}
			defer blobResp.Body.Close()

			// 将分片文件内容写入响应主体
			_, err = io.Copy(w, blobResp.Body)
			if err != nil {
				log.Println("写入响应主体数据时发生错误:", err)
				return
			}
		}
	} else {
		// 设置响应头，准备下载整个文件
		w.Header().Set("Content-Disposition", "inline; filename="+fileName)
		// 使用DetectContentType函数检测文件类型
		w.Header().Set("Content-Type", http.DetectContentType(buffer))
		// 将文件内容写入响应主体
		_, err = w.Write(buffer[:n])
		if err != nil {
			http.Error(w, "写入内容失败", http.StatusInternalServerError)
			log.Println(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Println(http.StatusInternalServerError)
			return
		}
	}
}


// Index 首页
func Index(w http.ResponseWriter, r *http.Request) {
	htmlPath := TemplatesPath + "/images.tmpl"
	if conf.Mode == "p" {
		htmlPath = TemplatesPath + "/files.tmpl"
	}
	file, err := assets.Templates.ReadFile(htmlPath)
	if err != nil {
		http.Error(w, "HTML file not found", http.StatusNotFound)
		return
	}
	// 读取头部模板
	headerFile, err := assets.Templates.ReadFile(TemplatesPath + "/header.tmpl")
	if err != nil {
		http.Error(w, "Header template not found", http.StatusNotFound)
		return
	}

	// 读取页脚模板
	footerFile, err := assets.Templates.ReadFile(TemplatesPath + "/footer.tmpl")
	if err != nil {
		http.Error(w, "Footer template not found", http.StatusNotFound)
		return
	}

	// 创建HTML模板并包括头部
	tmpl := template.New("html")
	tmpl, err = tmpl.Parse(string(headerFile))
	if err != nil {
		http.Error(w, "Error parsing header template", http.StatusInternalServerError)
		return
	}

	// 包括主HTML内容
	tmpl, err = tmpl.Parse(string(file))
	if err != nil {
		http.Error(w, "Error parsing HTML template", http.StatusInternalServerError)
		return
	}

	// 包括页脚
	tmpl, err = tmpl.Parse(string(footerFile))
	if err != nil {
		http.Error(w, "Error parsing footer template", http.StatusInternalServerError)
		return
	}

	// 直接将HTML内容发送给客户端
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering HTML template", http.StatusInternalServerError)
	}
}

func Pwd(w http.ResponseWriter, r *http.Request) {
	// 输出 HTML 表单
	if r.Method != http.MethodPost {
		file, err := assets.Templates.ReadFile(TemplatesPath + "/pwd.tmpl")
		if err != nil {
			http.Error(w, "HTML file not found", http.StatusNotFound)
			return
		}
		// 读取头部模板
		headerFile, err := assets.Templates.ReadFile(TemplatesPath + "/header.tmpl")
		if err != nil {
			http.Error(w, "Header template not found", http.StatusNotFound)
			return
		}

		// 创建HTML模板并包括头部
		tmpl := template.New("html")
		if tmpl, err = tmpl.Parse(string(headerFile)); err != nil {
			http.Error(w, "Error parsing Header template", http.StatusInternalServerError)
			return
		}

		// 包括主HTML内容
		if tmpl, err = tmpl.Parse(string(file)); err != nil {
			http.Error(w, "Error parsing File template", http.StatusInternalServerError)
			return
		}

		// 直接将HTML内容发送给客户端
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, "Error rendering HTML template", http.StatusInternalServerError)
		}
		return
	}
	// 设置cookie
	cookie := http.Cookie{
		Name:  "p",
		Value: r.FormValue("p"),
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 只有当密码设置并且不为"none"时，才进行检查
		if conf.Pass != "" && conf.Pass != "none" {
			if cookie, err := r.Cookie("p"); err != nil || cookie.Value != conf.Pass {
				http.Redirect(w, r, "/pwd", http.StatusSeeOther)
				return
			}
		}
		next(w, r)
	}
}
