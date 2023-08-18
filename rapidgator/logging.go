package rapidgator

import (
	"fmt"

	"github.com/mattn/go-colorable"
)

func Log(text string) {
	fmt.Fprintln(colorable.NewColorableStdout(), "\033[32m[RapidGator]\033[0m "+text)
}

func Log_Error(text string) {
	fmt.Fprintln(colorable.NewColorableStdout(), "\033[38;5;208m[ERROR] \033[0m\033[32m[RapidGator]\033[0m "+text)
}
