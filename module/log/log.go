/* 写入日志文件的zerologer */
package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
var logMutex sync.Mutex

// 日志解析字段顺序
var zeroPartsOrder = []string{
	zerolog.LevelFieldName,
	zerolog.TimestampFieldName,
	zerolog.CallerFieldName,
	zerolog.MessageFieldName,
}

var (
	logLevel       string = "info"   //日志等级
	rotateFileSize int64  = 10 << 20 //日志文件大小默认10MB
)

var file *os.File //文件句柄

func initFile() bool {
	// 检查日志文件是否存在
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 如果不存在，则创建日志文件
		if _, err := os.Create(logPath); err != nil {
			fmt.Printf("Failed to create log file: %v\n", err)
			return false
		}
		fmt.Printf("Log file created: %s\n", logPath)
	} else if err != nil {
		// 如果发生其他错误，则输出错误信息并退出程序
		fmt.Printf("Error checking log file: %v\n", err)
		return false
	} else {
		fmt.Printf("Log file ok: %s\n", logPath)
	}

	// 检查归档路径是否存在
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		// 如果不存在，则创建归档文件夹
		if err := os.MkdirAll(archivePath, 0755); err != nil {
			fmt.Printf("Failed to create archive directory: %v\n", err)
			return false
		}
		fmt.Printf("Archive directory created: %s\n", archivePath)
	} else if err != nil {
		// 如果发生其他错误，则输出错误信息并退出程序
		fmt.Printf("Error checking archive directory: %v\n", err)
		return false
	} else {
		fmt.Printf("Archive directory ok: %s\n", archivePath)
	}

	// 检查日志文件的权限
	info, err := os.Stat(logPath)
	if err != nil {
		fmt.Printf("Error checking log file permission: %v\n", err)
		return false
	}
	// 如果日志文件的权限不足，则修改权限为 0644
	if info.Mode().Perm() != 0644 {
		if err := os.Chmod(logPath, 0644); err != nil {
			fmt.Printf("Error changing log file permission: %v\n", err)
			return false
		}
		fmt.Printf("Log file permission changed to 0644: %s\n", logPath)
	} else {
		fmt.Printf("Log file permission is correct: %s\n", logPath)
	}
	return true
}

// 创建新的日志文件
func createLogFile() bool {
	file, err := os.Create(logPath)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	defer file.Close()

	fmt.Println("New log file created.")
	return true
}

// 关闭并重新创建日志文件
func closeAndCreateLogFile() bool {
	//先关闭文件
	output.Close()
	err := os.Rename(logPath, archivePath+"/"+logPath+"."+time.Now().Format("20060102150405"))
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	return createLogFile()
}

func updateLogger() bool {
	//打开文件
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	//跟新logger
	//TODO: 需要加锁？
	output.Out = f
	Logger = zerolog.New(output).With().Caller().Timestamp().Logger()

	//关闭之前的句柄，并记录新句柄
	file.Close()
	file = f
	return true
}

func checkFile() bool {
	logMutex.Lock()
	defer logMutex.Unlock()
	// 检查日志文件是否存在
	_, err := os.Stat(logPath)

	// 如果日志文件不存在
	if os.IsNotExist(err) {
		if createLogFile() {
			return updateLogger()
		} else {
			return false
		}
	}

	// 获取日志文件大小
	fileInfo, err := os.Stat(logPath)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	fileSize := fileInfo.Size()

	// 判断是否需要更新文件
	if fileSize > rotateFileSize {
		if closeAndCreateLogFile() {
			return updateLogger()
		} else {
			return false
		}
	}
	return true
}

func Init() {

	//TODO: 从配置文件获取日志文件的路径、归档路径、文件大小,判断参数有效性

	if !initFile() {
		os.Exit(1)
	}

	//打开文件
	var err error
	file, err = os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	output.Out = file

	//格式化输出
	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("%-5s", i))
	}
	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}
	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatCaller = func(i interface{}) string {
		return fmt.Sprintf("[%s]", i)
	}

	//短文件名风格
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	//初始化全局log.Logger
	Logger = zerolog.New(output).With().Caller().Timestamp().Logger()

	SetLevel(logLevel)

}

func SetLevel(level string) {
	if l, err := zerolog.ParseLevel(level); err != nil {
		//设置log level
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		Logger.Error().Err(err).Msg("logLevel config args invalid. set level=info.")
	} else {
		zerolog.SetGlobalLevel(l)
	}
}

// Output duplicates the global logger and sets w as its output.
func Output(w io.Writer) zerolog.Logger {
	checkFile()
	return Logger.Output(w)
}

// With creates a child logger with the field added to its context.
func With() zerolog.Context {
	checkFile()

	return Logger.With()
}

// Level creates a child logger with the minimum accepted level set to level.
func Level(level zerolog.Level) zerolog.Logger {
	checkFile()

	return Logger.Level(level)
}

// Sample returns a logger with the s sampler.
func Sample(s zerolog.Sampler) zerolog.Logger {
	checkFile()
	return Logger.Sample(s)
}

// Hook returns a logger with the h Hook.
func Hook(h zerolog.Hook) zerolog.Logger {
	checkFile()
	return Logger.Hook(h)
}

// Err starts a new message with error level with err as a field if not nil or
// with info level if err is nil.
//
// You must call Msg on the returned event in order to send the event.
func Err(err error) *zerolog.Event {
	checkFile()
	return Logger.Err(err)
}

// Trace starts a new message with trace level.
//
// You must call Msg on the returned event in order to send the event.
func Trace() *zerolog.Event {
	checkFile()
	return Logger.Trace()
}

func Debug() *zerolog.Event {
	checkFile()
	return Logger.Debug()
}

func Info() *zerolog.Event {
	checkFile()
	return Logger.Info()
}

func Warn() *zerolog.Event {
	checkFile()
	return Logger.Warn()
}

func Error() *zerolog.Event {
	checkFile()
	return Logger.Error()
}

func Fatal() *zerolog.Event {
	checkFile()
	return Logger.Fatal()
}

func Panic() *zerolog.Event {
	checkFile()
	return Logger.Panic()
}

func WithLevel(level zerolog.Level) *zerolog.Event {
	checkFile()
	return Logger.WithLevel(level)
}

func Log() *zerolog.Event {
	checkFile()
	return Logger.Log()
}

func Print(v ...interface{}) {
	checkFile()
	Logger.Debug().CallerSkipFrame(1).Msg(fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	checkFile()
	Logger.Debug().CallerSkipFrame(1).Msgf(format, v...)
}

func Ctx(ctx context.Context) *zerolog.Logger {
	checkFile()
	return zerolog.Ctx(ctx)
}
