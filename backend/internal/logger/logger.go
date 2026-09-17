package logger

import "log"

func Debug(v ...any) { log.Println(append([]any{"[debug]"}, v...)...) }
func Info(v ...any)  { log.Println(append([]any{"[info]"}, v...)...) }
func Warn(v ...any)  { log.Println(append([]any{"[warn]"}, v...)...) }
func Error(v ...any) { log.Println(append([]any{"[error]"}, v...)...) }
