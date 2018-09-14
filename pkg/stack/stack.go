package stack

import (
	"runtime"
	"strings"
)

// 输出堆栈信息, skip表示需要跳过的帧数
func Backtrace(skip int) string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	buf = buf[:n]

	s := string(buf)

	// 需要跳过的行数 = 帧数 * 2 + 1
	lines := skip*2 + 1
	count := 0
	index := strings.IndexFunc(s, func(c rune) bool {
		if c != '\n' {
			return false
		}
		count++
		return count == lines
	})
	return s[index+1:]
}
