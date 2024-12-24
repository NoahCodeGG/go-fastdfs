package server

import (
	"bytes"
	log "github.com/sjqzhang/seelog"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"os/exec"
)

func (c *Server) GetVideoCoverByBytes(w http.ResponseWriter, data []byte, width, height uint) {
	tmpVideo, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		log.Error("创建临时视频文件失败:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpVideo.Name())
	defer tmpVideo.Close()

	if _, err = tmpVideo.Write(data); err != nil {
		log.Error("写入视频数据失败:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpImage, err := os.CreateTemp("", "cover-*.jpg")
	if err != nil {
		log.Error("创建临时图片文件失败:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpImage.Name())
	defer tmpImage.Close()

	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", tmpVideo.Name(),
		"-ss", "1",
		"-vframes", "1",
		"-f", "image2",
		tmpImage.Name())

	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		log.Error("执行ffmpeg命令失败:", err, "输出:", string(output))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	coverData, err := os.ReadFile(tmpImage.Name())
	if err != nil {
		log.Error("读取生成的封面失败:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if width > 0 && height > 0 {
		buf := bytes.NewBuffer(coverData)
		c.ResizeImageByBytes(w, buf.Bytes(), width, height)
	} else {
		var (
			img     image.Image
			imgType string
			reader  = bytes.NewReader(coverData)
		)

		img, imgType, err = image.Decode(reader)
		if err != nil {
			log.Error(err)
			return
		}

		if imgType == "jpg" || imgType == "jpeg" {
			jpeg.Encode(w, img, nil)
		} else if imgType == "png" {
			png.Encode(w, img)
		} else {
			w.Write(data)
		}
	}
}

func (c *Server) GetVideoCover(w http.ResponseWriter, fullpath string, width, height uint) {
	data, err := os.ReadFile(fullpath)
	if err != nil {
		log.Error("读取视频文件失败:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.GetVideoCoverByBytes(w, data, width, height)
}
