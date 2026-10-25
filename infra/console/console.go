// Package console 负责面向用户的终端输出，与结构化日志分开。
package console

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type Console struct {
	out io.Writer
}

func New() *Console {
	return &Console{out: os.Stdout}
}

func (vi *Console) Println(msg string) {
	fmt.Fprintln(vi.out, msg)
}

func (vi *Console) Printf(format string, args ...any) {
	fmt.Fprintf(vi.out, format, args...)
}

// Info 以绿色加粗输出一行带 [INFO] 前缀的提示。
func (vi *Console) Info(msg string) {
	c := color.New(color.FgGreen, color.Bold)
	c.Fprintf(vi.out, "[INFO] %s\n", msg)
}
