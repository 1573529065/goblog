package logger

import "log"

// logError 记录错误
func LogError(err error) {
	if err != nil {
		log.Println(err)
	}
}
