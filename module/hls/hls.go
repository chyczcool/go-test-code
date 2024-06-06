// 实现一个支持在线点播回放的hls服务
package hls

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grafov/m3u8"
)

var (
	hlsSrv = http.Server{}
	ctx    = context.Background()

	//保存切片的临时文件夹
	//切片路径和播放列表生成规则：$(HLS_TEMP_DIR)/<VIDEO_FILE_NAME>/<START_TIME>-<END_TIME>/*.ts
	HLS_TEMP_DIR = filepath.Join(os.TempDir(), "mediagate")
)

const (
	HLS_PORT = ":8091"

	HLS_MEDIA_URI          = "/media/hls/stream.m3u8"
	HLS_MEDIA_PLAYLIST_URI = "/media/hls/playlist.m3u8"
	HLS_SEGMENT_TS_URI     = "/media/hls/segment.ts"

	SEGMENT_TIME_PARTICLE        = 5                         //分片片时间粒度，单位为秒
	MEDIA_PLAYLIST_TIME_PARTICLE = 2 * SEGMENT_TIME_PARTICLE //播放列表时间长度

	HLS_MEDIA_PLAYLIST_FILE_NAME = "playlist.m3u8"
)

// GetMeiaFileDuration 获取媒体文件长度
func GetMeiaFileDuration(filePath string) (duration float64, err error) {
	ffprobeCmd := exec.Command("ffprobe", "-v", "quiet", "-show_format", "-of", "default=noprint_wrappers=1:nokey=0", "-i", filePath)

	ffprobeOut, err := ffprobeCmd.StdoutPipe()
	if err != nil {
		fmt.Println("ffprobe stdoutpipe err.")
		return 0, err
	}

	if err := ffprobeCmd.Start(); err != nil {
		fmt.Println("ffprobe Start err.")
		return 0, err
	}

	grepCmd := exec.Command("grep", "duration")
	grepCmd.Stdin = ffprobeOut

	grepOut, err := grepCmd.StdoutPipe()
	if err != nil {
		fmt.Println("grepCmd StdoutPipe err.")
		return 0, err
	}

	if err := grepCmd.Start(); err != nil {
		fmt.Println("grepCmd Start err.")
		return 0, err
	}

	cutCmd := exec.Command("cut", "-d", "=", "-f", "2")
	cutCmd.Stdin = grepOut

	var cutOut bytes.Buffer
	cutCmd.Stdout = &cutOut

	if err := cutCmd.Start(); err != nil {
		fmt.Println("cutCmd Start err.")
		return 0, err
	}

	if err := ffprobeCmd.Wait(); err != nil {
		fmt.Println("ffprobeCmd Wait err.")
		return 0, err
	}

	if err := grepCmd.Wait(); err != nil {
		fmt.Println("grepCmd Wait err.")
		return 0, err
	}
	if err := cutCmd.Wait(); err != nil {
		fmt.Println("cutCmd Wait err.")
		return 0, err
	}

	return strconv.ParseFloat(strings.TrimSpace(cutOut.String()), 64)
}

// Start 开始一个http服务
func Start() {
	go func() {
		if err := hlsSrv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
}

// Shutdown 关闭http服务
func Shutdown() {
	hlsSrv.Shutdown(ctx)
}

// handlerHLSStreaming 处理hls流媒体请求,返回master list
func handlerHLSStreaming(rsp http.ResponseWriter, req *http.Request) {
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.Header().Set("Content-Type", "application/vnd.apple.mpegurl")

	if req.Method == "OPTIONS" {
		rsp.Header().Set("Access-Control-Allow-Methods", "GET")
		return
	}

	//获取url参数
	src := req.URL.Query().Get("src")
	if src == "" {
		http.Error(rsp, "no URL provided", http.StatusBadRequest)
		return
	}

	//TODO:判断媒体资源是否存在

	//创建会话
	Session, err := NewSession(src, VIDEO_CODEC_COPY, true)
	if err != nil {
		http.Error(rsp, "failed to create session: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//如果是点播模式，在处理master请求时就切片整个文件
	if Session.isVod {
		playlist := filepath.Join(Session.TempRoot, Session.videoCodec, Session.mediaPlaylistFilename)
		if _, err = os.Stat(playlist); os.IsNotExist(err) {
			Session.CreatePlaylist()
		}
	}

	//记录
	key, ok := Sessions.Set(Session)
	if !ok {
		http.Error(rsp, "failed to create session", http.StatusInternalServerError)
		return
	}

	//构建master list
	m := m3u8.NewMasterPlaylist()
	p, _ := m3u8.NewMediaPlaylist(5, 10)
	uri := HLS_MEDIA_PLAYLIST_URI + fmt.Sprintf("?src=%s", key)
	m.Append(uri, p, m3u8.VariantParams{Bandwidth: 192000, Codecs: "avc1.640029,fLaC"})

	//http响应master playlist
	rsp.Write([]byte(m.String()))
}

// handlerHLSPlaylist 处理hls播放列表请求,返回play list
func handlerHLSPlaylist(rsp http.ResponseWriter, req *http.Request) {
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.Header().Set("Content-Type", "application/vnd.apple.mpegurl")

	if req.Method == "OPTIONS" {
		rsp.Header().Set("Access-Control-Allow-Methods", "GET")
		return
	}

	//get Query src参数
	src := req.URL.Query().Get("src")
	if src == "" {
		http.Error(rsp, "missing required parameter src", http.StatusBadRequest)
		return
	}

	//获取session
	s, ok := Sessions[src]
	if !ok {
		http.Error(rsp, "session not found", http.StatusNotFound)
		return
	}

	var playProgressUint uint64 = 0
	var err error
	var isJump bool = false
	//获取请求头 vod_play_progress
	playProgress := req.Header.Get("vod_play_progress")
	if playProgress != "" {
		playProgressUint, err = strconv.ParseUint(playProgress, 10, 64)
		if err != nil {
			http.Error(rsp, fmt.Sprintf("invalid header vod_play_progress: %s", playProgress), http.StatusBadRequest)
			return
		}
		isJump = true
	}

	pl, err := s.GetMediaPlaylist(playProgressUint, isJump)
	if err != nil {
		if err == ErrSessionWaitSec {
			//返回一个空的播放列表
			rsp.Write(pl)
		} else {
			http.Error(rsp, fmt.Sprintf("get media playlist error: %s", err), http.StatusInternalServerError)
		}
		return
	}

	//pl写入rsp
	if wl, err := rsp.Write(pl); err != nil {
		//TODO: 日志记录错误
		_ = wl
	}
}

// handlerHLSSegment 处理hls片段请求,返回segment
func handlerHLSSegmentTS(rsp http.ResponseWriter, req *http.Request) {
	rsp.Header().Set("Access-Control-Allow-Origin", "*")
	rsp.Header().Set("Content-Type", "application/vnd.apple.mpegurl")

	if req.Method == "OPTIONS" {
		rsp.Header().Set("Access-Control-Allow-Methods", "GET")
		return
	}

	//获取file参数
	segment := req.URL.Query().Get("file")
	if segment == "" {
		http.Error(rsp, "invalid file", http.StatusBadRequest)
		return
	}

	segmentRealPath := filepath.Join(HLS_TEMP_DIR, segment)

	file, err := os.Open(segmentRealPath)
	if err != nil {
		http.Error(rsp, fmt.Sprintf("open file error: %s", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(rsp, fmt.Sprintf("stat file error: %s", err), http.StatusInternalServerError)
		return
	}

	// 设置响应头
	rsp.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	rsp.Header().Set("Content-Type", "application/octet-stream")
	rsp.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	http.ServeContent(rsp, req, fileInfo.Name(), fileInfo.ModTime(), file)
}

// Init 初始化一个HLS服务器
func Init() {
	//绑定路由
	handler := http.NewServeMux()
	handler.HandleFunc(HLS_MEDIA_URI, handlerHLSStreaming)
	handler.HandleFunc(HLS_MEDIA_PLAYLIST_URI, handlerHLSPlaylist)
	handler.HandleFunc(HLS_SEGMENT_TS_URI, handlerHLSSegmentTS)

	hlsSrv.Addr = fmt.Sprintf(":%d", 12345)
	hlsSrv.Handler = handler

	//TODO:启动一个任务去定时删除超时session
	//TODO:启动一个任务去定时清理切片文件

	go Start()
}
