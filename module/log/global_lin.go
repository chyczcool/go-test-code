//go:build linux

package log

import (
	"time"

	"github.com/rs/zerolog"
)

var (
	logPath     string = "/var/log/media_gate.log" //日志文件
	archivePath string = "/var/log/backup"         //归档路径
)

var output = zerolog.ConsoleWriter{
	TimeFormat: time.DateTime,
	PartsOrder: zeroPartsOrder,
}
