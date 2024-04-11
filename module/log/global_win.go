//go:build windows

package log

import (
	"time"

	"github.com/rs/zerolog"
)

var (
	logPath     string = "./media_gate.log" //日志文件
	archivePath string = "./backup"         //归档路径
)

// win 关闭颜色输出
var output = zerolog.ConsoleWriter{
	TimeFormat: time.DateTime,
	PartsOrder: zeroPartsOrder,
	NoColor:    true,
}
