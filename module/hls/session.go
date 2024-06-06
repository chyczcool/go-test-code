package hls

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafov/m3u8"
)

var ErrSessionWaitSec error = fmt.Errorf("wait a second")

var (
	Sessions SessionMap = make(map[string]*Session)

	SessionId      uint64 = 0
	SessionIdMutex sync.Mutex
)

type SessionMap map[string]*Session

func (s SessionMap) Get(key string) *Session {
	return s[key]
}

func hashMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func (s SessionMap) Set(session *Session) (key string, ok bool) {
	if session == nil {
		return "", false
	}

	// 3次尝试
	for i := 0; i < 3; i++ {
		key = hashMD5(fmt.Sprintf("%d:%d:%s", session.Id, time.Now().Unix(), session.FilePath))
		if !s.ContainsKey(key) {
			s[key] = session
			return key, true
		} else {
			continue
		}
	}
	return "", false
}

func (s SessionMap) Del(key string) {
	delete(s, key)
}

func (s SessionMap) Len() int {
	return len(s)
}

func (s SessionMap) Keys() []string {
	keys := make([]string, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

// ContainsKey 判断是否包含id
func (s SessionMap) ContainsKey(id string) bool {
	_, ok := s[id]
	return ok
}

const (
	SESSION_VOD_NOT_LAUNCHED = iota + 1
	SESSION_VOD_LAUNCHED
	SESSION_VOD_COMPLETED
)

type Session struct {
	Id       uint64 //自增id
	FilePath string //session指向的文件路径
	Sequence uint64 //当前播放的序列号
	TempRoot string //临时文件夹

	PlayProgress uint64  //播放进度
	Duration     float64 //总时长
	IsEndList    bool    //结束播放

	mediaPlaylistTimeParticle uint16 //媒体播放列表的粒度
	mediaPlaylistFilename     string //播放列表文件名
	segmentTimeParticle       uint16 //切片粒度
	videoCodec                string //视频编码格式

	isVod     bool //是否为点播模式
	vodStatus int8 //点播列表状态
	hasErr    bool
	vodErr    error //vod切片返回的错误
	vodPp     *m3u8.MediaPlaylist

	vodHlsBaseUrl string
}

// NewSession 创建一个session
func NewSession(inputFile string, vcodec string, isVod bool) (*Session, error) {
	SessionIdMutex.Lock()
	SessionId++
	SessionIdMutex.Unlock()

	duration, err := GetMeiaFileDuration(inputFile)
	if err != nil {
		return nil, err
	}

	//创建文件临时文件夹，/tmp/mediagate/demo.mp4/
	tmpDir := filepath.Join(HLS_TEMP_DIR, inputFile)
	err = os.MkdirAll(tmpDir, 0755)
	if err != nil {
		return nil, err
	}

	return &Session{
		Id:                        SessionId,
		FilePath:                  inputFile,
		Sequence:                  0,
		PlayProgress:              0,
		Duration:                  duration,
		TempRoot:                  tmpDir,
		IsEndList:                 false,
		mediaPlaylistTimeParticle: MEDIA_PLAYLIST_TIME_PARTICLE,
		mediaPlaylistFilename:     HLS_MEDIA_PLAYLIST_FILE_NAME,
		segmentTimeParticle:       SEGMENT_TIME_PARTICLE,
		videoCodec:                vcodec,

		isVod:         isVod,
		vodStatus:     SESSION_VOD_NOT_LAUNCHED,
		hasErr:        false,
		vodErr:        nil,
		vodHlsBaseUrl: fmt.Sprintf("%s?file=%s/%s/", HLS_SEGMENT_TS_URI, inputFile, vcodec),
	}, nil
}

// ParseMediaPlaylsitSequence 解析文件segment数量
func ParseMediaPlaylistSegmentCount(file string) (uint64, error) {
	f, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		return 0, err
	}

	pp, ok := p.(*m3u8.MediaPlaylist)
	if !ok {
		return 0, fmt.Errorf("media play list format err")
	}

	return uint64(len(pp.Segments)), nil
}

// 视频编码参数
const (
	VIDEO_CODEC_264  = "libx264"
	VIDEO_CODEC_COPY = "copy"
)

// GetFFmpegArg 获取ffmpeg参数
func (s *Session) GetFFmpegArg(segmentDir string, isVod bool) []string {
	to := s.PlayProgress + uint64(s.mediaPlaylistTimeParticle)
	templatePath := filepath.Join(segmentDir, `segment_%02d.ts`)
	playlist := filepath.Join(segmentDir, s.mediaPlaylistFilename)
	segmentBaseUrl := fmt.Sprintf("%s?file=%s/%d-%d/", HLS_SEGMENT_TS_URI, s.FilePath, s.PlayProgress, to)
	forceKeyFrame := fmt.Sprintf("expr:gte(t,n_forced*%d)", s.segmentTimeParticle)

	//回放模式
	if isVod {
		return []string{
			"-hide_banner",
			"-i", s.FilePath,
			"-c:v", s.videoCodec,
			"-c:a", "copy",
			"-hls_base_url", s.vodHlsBaseUrl,
			"-hls_segment_filename", templatePath,
			"-hls_time", strconv.FormatUint(uint64(s.segmentTimeParticle), 10),
			"-f", "hls",
			"-hls_playlist_type", "vod",
			playlist,
			"-y",
		}
	}

	if !s.IsEndList {
		return []string{
			"-hide_banner",
			"-ss", strconv.FormatUint(s.PlayProgress, 10),
			"-i", s.FilePath,
			"-c:a", "copy", //指定音频编码方式，默认是copy按需编码
			"-c:v", s.videoCodec, //指定视频编码方式，默认是libx264按需编码
			"-force_key_frames", forceKeyFrame, //在每个切片前插入关键帧
			"-hls_base_url", segmentBaseUrl, //设置segment的url前缀
			"-hls_list_size", "0", //保留所有生成的切片
			"-hls_segment_filename", templatePath, //设置切片文件命名规则
			"-hls_time", strconv.FormatUint(uint64(s.segmentTimeParticle), 10), //设置输出切片的长度，如果c:v copy则可能不会生效
			"-preset", "ultrafast", //快速编码，会降低编码质量
			"-copyts",
			"-start_at_zero",                                    //重置输出文件的时间戳，使其从零开始计时
			"-start_number", strconv.FormatUint(s.Sequence, 10), //设置切片开始sequence
			"-t", strconv.FormatUint(to, 10), //处理到to为止
			"-f", "hls", //指定输出格式，默认是hls
			"-hls_flags", "omit_endlist", //如果视频结束，设置end list
			playlist,
			"-y", //覆盖原始文件
		}
	} else {
		return []string{
			"-hide_banner",
			"-ss", strconv.FormatUint(s.PlayProgress, 10),
			"-i", s.FilePath,
			"-c:a", "copy", //指定音频编码方式，默认是copy按需编码
			"-c:v", s.videoCodec, //指定视频编码方式，默认是libx264按需编码
			"-force_key_frames", forceKeyFrame, //在每个切片前插入关键帧
			"-hls_base_url", segmentBaseUrl, //设置segment的url前缀
			"-hls_list_size", "0", //保留所有生成的切片
			"-hls_segment_filename", templatePath, //设置切片文件命名规则
			"-hls_time", strconv.FormatUint(uint64(s.segmentTimeParticle), 10), //设置输出切片的长度，如果c:v copy则可能不会生效
			"-preset", "ultrafast", //快速编码，会降低编码质量
			"-copyts",                                           //保留原始输入时间戳,而不是直接生成
			"-start_at_zero",                                    //重置输出文件的时间戳，使其从零开始计时
			"-start_number", strconv.FormatUint(s.Sequence, 10), //设置切片开始sequence
			"-t", strconv.FormatUint(to, 10), //处理到to为止
			"-f", "hls", //指定输出格式，默认是hls
			playlist,
			"-y", //覆盖原始文件
		}
	}
}

func (s *Session) CreateVodPlaylistRun(cmd *exec.Cmd) {
	s.vodStatus = SESSION_VOD_LAUNCHED

	err := cmd.Run()
	if err != nil {
		s.hasErr = true
		if e, ok := cmd.Stderr.(fmt.Stringer); ok {
			s.vodErr = fmt.Errorf("%s", e)
		} else {
			s.vodErr = fmt.Errorf("%s", err)
		}
	}
	s.vodStatus = SESSION_VOD_COMPLETED

	meidiaPlaylist := filepath.Join(s.TempRoot, s.videoCodec, s.mediaPlaylistFilename)

	//解析playlist文件
	f, err := os.Open(meidiaPlaylist)
	if err != nil {
		s.hasErr = true
		s.vodErr = fmt.Errorf("%s", err)
		return
	}
	defer f.Close()

	p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		s.hasErr = true
		s.vodErr = fmt.Errorf("%s", err)
		return
	}
	var ok bool
	s.vodPp, ok = p.(*m3u8.MediaPlaylist)
	if !ok {
		s.hasErr = true
		s.vodErr = fmt.Errorf("not a media playlist")
		return
	}
}

// CreatePlaylist 创建媒体播放列表m3u8文件
func (s *Session) CreatePlaylist() error {
	//生成m3u8文件
	//NOTE:切片路径和播放列表生成规则：$(HLS_TEMP_DIR)/<VIDEO_FILE_NAME>/<START_TIME>-<END_TIME>/*.ts
	segmentDir := filepath.Join(s.TempRoot, fmt.Sprintf("%d-%d", s.PlayProgress, s.PlayProgress+uint64(s.mediaPlaylistTimeParticle)))

	//如果是vod模式，则切片路径为$(HLS_TEMP_DIR)/<VIDEO_CODEC>/*.ts
	if s.isVod {
		segmentDir = filepath.Join(s.TempRoot, s.videoCodec)
	}

	//创建segment目录
	if err := os.MkdirAll(segmentDir, 0755); err != nil {
		return err
	}

	args := s.GetFFmpegArg(segmentDir, s.isVod)
	cmd := exec.Command("ffmpeg", args...)

	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	fmt.Println("exec:", cmd.String())

	if s.isVod {
		//异步生成vod切片
		//TODO:增加信号机制，解决第一次播放新文件时，media playlist未来得及生成播放失败的情况
		go s.CreateVodPlaylistRun(cmd)
	} else {
		//如果是实时切片，阻塞执行ffmpeg切片命令
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("%s", stderrBuf.String())
		}
	}

	return nil
}

func (s *Session) RevisePlaylist(pp *m3u8.MediaPlaylist) (*m3u8.MediaPlaylist, error) {
	//如果是libx264编码，一般在第一个关键帧时，会先生成一个0长度的segment
	if s.videoCodec == VIDEO_CODEC_264 {
		segments := pp.GetAllSegments()
		if segments == nil {
			return nil, fmt.Errorf("no segments")
		}
		//NOTE:如果其他位置出现了0长度的segment，则需要遍历segments，重新new一个MediaPlaylist返回
		if segments[0].Duration < 0.1 {
			pp.Remove()
		}
	}

	//设置playlist的sequence，涉及点播时，需要修改之前生成的切片的sequence
	if pp.SeqNo != s.Sequence {
		pp.SeqNo = s.Sequence
	}

	//更新sequence
	s.Sequence += uint64(pp.Count())

	return pp, nil
}

func (s *Session) SetInitMp4(pp *m3u8.MediaPlaylist) error {
	//将第一个segment设置为init_mp4
	segments := pp.GetAllSegments()

	initSegment := segments[0]

	//只有一个file参数
	inputSegment := filepath.Join(HLS_TEMP_DIR, strings.Split(initSegment.URI, "=")[1])
	segmentDir := filepath.Dir(inputSegment)
	outputInitMp4 := filepath.Join(segmentDir, "init.mp4")

	args := []string{
		"-hide_banner",
		"-i", inputSegment,
		"-c:v", "copy",
		"-c:a", "copy",
		outputInitMp4,
		"-y",
	}

	cmd := exec.Command("ffmpeg", args...)

	stdoutBuf := bytes.NewBuffer(nil)
	stderrBuf := bytes.NewBuffer(nil)
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s", stderrBuf.String())
	}

	time.Sleep(time.Millisecond * 500)

	initUri := fmt.Sprintf("%s?file=%s", HLS_SEGMENT_TS_URI, strings.TrimPrefix(outputInitMp4, HLS_TEMP_DIR))

	pp.SetMap(initUri, 0, 0)

	//移除第一个segment
	pp.Remove()
	return nil
}

// GetVodMediaPlaylsitByPath 获取vod的playlist
func (s *Session) GetVodMediaPlaylsitByPath(mediaPalyList string) (mp []byte, err error) {
	if s.hasErr {
		return nil, s.vodErr
	}

	pp, err := m3u8.NewMediaPlaylist(10, 10)
	if err != nil {
		return nil, err
	}
	pp.SeqNo = s.Sequence

	var vodPp *m3u8.MediaPlaylist
	//切片完成
	if s.vodStatus == SESSION_VOD_COMPLETED {
		vodPp = s.vodPp
	} else {
		//切片没完成时需要实时解析playlist
		//解析playlist文件
		f, err := os.Open(mediaPalyList)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), false)
		if err != nil {
			return nil, err
		}
		var ok bool
		vodPp, ok = p.(*m3u8.MediaPlaylist)
		if !ok {
			return nil, err
		}
	}

	segments := vodPp.GetAllSegments()

	//跳到s.playProgress位置
	var pregresss float64
	var i uint
	for i = 0; i < vodPp.Count(); i++ {
		if pregresss >= float64(s.PlayProgress) {
			break
		}
		pregresss += segments[i].Duration
	}

	if i == vodPp.Count() {
		if s.vodStatus == SESSION_VOD_COMPLETED {
			return nil, fmt.Errorf("media over")
		} else {
			return []byte(pp.String()), ErrSessionWaitSec
		}
	}

	//装填即将返回的
	var playlistDuration float64
	var playlistCount uint
	for j := i; j < vodPp.Count(); j++ {
		if playlistDuration >= float64(s.mediaPlaylistTimeParticle) {
			break
		}

		pp.AppendSegment(segments[j])

		playlistDuration += segments[j].Duration
		playlistCount++
	}

	s.PlayProgress += uint64(playlistDuration)

	s.Sequence += uint64(playlistCount)

	//设置结束标签
	if s.IsEndList {
		pp.Closed = true
	}

	return []byte(pp.String()), nil
}

// GetMediaPlaylistByPath 获取播放列表
func (s *Session) GetMediaPlaylistByPath(mediaPalyList string) (mp []byte, err error) {

	//如果不存在播放列表，则创建一个
	if _, err := os.Stat(mediaPalyList); os.IsNotExist(err) {
		if err := s.CreatePlaylist(); err != nil {
			return []byte{}, err
		}
	}

	f, err := os.Open(mediaPalyList)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()

	p, _, err := m3u8.DecodeFrom(bufio.NewReader(f), false)
	if err != nil {
		return []byte{}, err
	}

	pp, ok := p.(*m3u8.MediaPlaylist)
	if !ok {
		return []byte{}, fmt.Errorf("media play list format err")
	}

	//检查可能的错误
	s.RevisePlaylist(pp)

	//增加进度条
	s.PlayProgress += uint64(s.mediaPlaylistTimeParticle)

	return pp.Encode().Bytes(), nil
}

// GetMediaPlaylist 获取media playlist
func (s *Session) GetMediaPlaylist(requestTimePoint uint64, isJump bool) ([]byte, error) {
	if isJump {
		if requestTimePoint > uint64(s.Duration) {
			return nil, fmt.Errorf("req time bigger than duration")
		}

		//切片对齐
		s.PlayProgress = uint64(s.mediaPlaylistTimeParticle)*(requestTimePoint/uint64(s.mediaPlaylistTimeParticle)) +
			uint64(s.mediaPlaylistTimeParticle)
	}

	//是否为最后一个播放列表
	if s.PlayProgress+uint64(s.mediaPlaylistTimeParticle) > uint64(s.Duration) {
		s.IsEndList = true
	}

	segmentDir := filepath.Join(s.TempRoot, fmt.Sprintf("%d-%d", s.PlayProgress, s.PlayProgress+uint64(s.mediaPlaylistTimeParticle)))
	if s.isVod {
		segmentDir = filepath.Join(s.TempRoot, s.videoCodec)
	}
	mediaPalyList := filepath.Join(segmentDir, s.mediaPlaylistFilename)

	if s.isVod {
		//TODO:回放模式的playlist 返回逻辑修改成尽可能多的返回可访问的切片，即当ffmpeg任务结束时就返回整个或剩下的所有切片列表，解决回放时没有播放进度的问题
		return s.GetVodMediaPlaylsitByPath(mediaPalyList)
	} else {
		return s.GetMediaPlaylistByPath(mediaPalyList)
	}
}
