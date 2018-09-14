// +build debug,!disable_color

package colorized

func Blue(str string) string    { return "\x1b[0;34m" + str + "\x1b[0m" }
func Yellow(str string) string  { return "\x1b[0;33m" + str + "\x1b[0m" }
func Green(str string) string   { return "\x1b[0;32m" + str + "\x1b[0m" }
func Magenta(str string) string { return "\x1b[0;35m" + str + "\x1b[0m" }
func Cyan(str string) string    { return "\x1b[0;36m" + str + "\x1b[0m" }
func Gray(str string) string    { return "\x1b[0;37m" + str + "\x1b[0m" }
func White(str string) string   { return "\x1b[0;30m" + str + "\x1b[0m" }
