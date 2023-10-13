package utils

import (
	"fmt"
	"github.com/logrusorgru/aurora/v4"
)

func Info(msg string) {
	fmt.Printf(
		"%s %s\n",
		aurora.Bold(aurora.BgBlue("INFO")),
		aurora.BrightBlue(msg),
	)
}

func Log(msg string) {
	fmt.Printf(
		"%s %s\n",
		aurora.Bold(aurora.BgGray(7, "LOG")),
		aurora.Gray(14, msg),
	)
}

func Warning(msg string) {
	fmt.Printf(
		"%s %s\n",
		aurora.Bold(aurora.BgYellow("WARNING")),
		aurora.BrightYellow(msg),
	)
}

func Error(err error) {
	fmt.Printf(
		"%s %s\n\n%s\n%s\n%s\n\n",
		aurora.Bold(aurora.BgRed("ERROR")),
		aurora.BrightRed(err),
		aurora.Bold("Need help? Here are some helpful links"),
		aurora.Blue("https://codigo.ai/community").Hyperlink("https://codigo.ai/community"),
		aurora.Blue("https://docs.codigo.ai").Hyperlink("https://docs.codigo.ai"),
	)
}
