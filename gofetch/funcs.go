package gofetch

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"strings"
	"unicode/utf8"
)

func tmpl_Shell(cmd string) string {
	call := exec.Command("sh", "-c", cmd)
	stdoutPipe, _ := call.StdoutPipe()
	if err := call.Start(); err != nil {
		log.Fatalf("cmd.Start() failed with %s\n", err)
	}

	stdoutChan := make(chan []byte)

	go func() {
		stdout := []byte{}
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if err != nil {
				break
			}
			stdout = append(stdout, buf[:n]...)
		}
		stdoutChan <- stdout
	}()

	err := call.Wait()
	if err != nil && err.Error() != "signal: killed" {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	return string(<-stdoutChan)
}

func tmpl_PadLeft(str string, char rune, width int) string {
	stripped := ansi.ReplaceAllString(str, "")
	l := utf8.RuneCountInString(stripped)
	return str + strings.Repeat(string(char), width-l)
}

func tmpl_PadRight(str string, char rune, width int) string {
	stripped := ansi.ReplaceAllString(str, "")
	l := utf8.RuneCountInString(stripped)
	return strings.Repeat(string(char), width-l) + str
}

func tmpl_PadCenter(str string, char rune, width int) string {
	stripped := ansi.ReplaceAllString(str, "")
	sz := utf8.RuneCountInString(stripped)
	w := float64(width - sz)
	left := int(math.Floor(w / 2))
	right := int(math.Ceil(w / 2))
	return strings.Repeat(string(char), left) + str + strings.Repeat(string(char), right)
}

func tmpl_HumanSize(pow2 bool, sz int) string {
	bases := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	ratio := 1024.0
	if !pow2 {
		bases = []string{"B", "KB", "MB", "GB", "TB", "PB"}
		ratio = 1000.0
	}
	f := float64(sz)
	i := 0
	for i < len(bases) && f > ratio {
		i++
		f /= ratio
	}
	return fmt.Sprintf("%.2f %s", f, bases[i])
}

func tmpl_Yank(key string, val any, obj []map[string]any) map[string]any {
	for _, o := range obj {
		if o[key] == val {
			return o
		}
	}
	return nil
}

func tmpl_AtMap(i int, arr []map[string]any) map[string]any {
	return arr[i]
}

func tmpl_At(i int, arr []any) any {
	return arr[i]
}

func tmpl_Key(i string, arr map[string]any) any {
	return arr[i]
}
