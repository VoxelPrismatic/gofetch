package gofetch

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/Masterminds/sprig/v3"
)

type Schema struct {
	Sections      []*Section
	Globals       *map[string]any
	ModuleFormat  *string
	SectionFormat *SecFormat
}

type SecFormat struct {
	Header *string
	Footer *string
}

type Module struct {
	Locals *map[string]any
	Key    *string
	Val    *string
	Format *string
}

type Section struct {
	Title         string
	Modules       []*Module
	ModuleFormat  *string
	SectionFormat *SecFormat
}

func parseMap(obj *map[string]any) {
	for k, v := range *obj {
		if r := reflect.TypeOf(v); r.Kind() == reflect.Map {
			m := v.(map[string]any)
			if m["Cast"] == nil {
				continue
			}

			var err error
			x := fmt.Sprint(m["Value"])
			switch m["Cast"] {
			case "int64":
				(*obj)[k], err = strconv.ParseInt(x, 0, 64)
			case "int32":
				var i int64
				i, err = strconv.ParseInt(x, 0, 32)
				(*obj)[k] = int32(i)
			case "int16":
				var i int64
				i, err = strconv.ParseInt(x, 0, 16)
				(*obj)[k] = int16(i)
			case "int8":
				var i int64
				i, err = strconv.ParseInt(x, 0, 8)
				(*obj)[k] = int8(i)
			case "int":
				var i int64
				i, err = strconv.ParseInt(x, 0, 64)
				(*obj)[k] = int(i)
			case "float64":
				(*obj)[k], err = strconv.ParseFloat(x, 64)
			case "float32":
				var i float64
				i, err = strconv.ParseFloat(x, 32)
				(*obj)[k] = float32(i)
			case "complex128":
				(*obj)[k], err = strconv.ParseComplex(x, 128)
			case "complex64":
				var i complex128
				i, err = strconv.ParseComplex(x, 64)
				(*obj)[k] = complex64(i)
			case "uint64":
				(*obj)[k], err = strconv.ParseUint(x, 0, 64)
			case "uint32":
				var i uint64
				i, err = strconv.ParseUint(x, 0, 32)
				(*obj)[k] = uint32(i)
			case "uint16":
				var i uint64
				i, err = strconv.ParseUint(x, 0, 16)
				(*obj)[k] = uint16(i)
			case "uint8":
				var i uint64
				i, err = strconv.ParseUint(x, 0, 8)
				(*obj)[k] = uint8(i)
			case "uint":
				var i uint64
				i, err = strconv.ParseUint(x, 0, 64)
				(*obj)[k] = uint(i)
			default:
				panic(fmt.Sprintf("cast not implemented: %s", m["Cast"]))
			}
			if err != nil {
				panic(err)
			}
			continue
		}
		if r := reflect.TypeOf(v); r.Kind() != reflect.String {
			continue
		}
		(*obj)[k] = parseTemplate(fmt.Sprint(v), nil)
	}
}

func (sch Schema) Parse() []string {
	if sch.Globals != nil {
		parseMap(sch.Globals)
	}

	ret := []string{}
	for _, section := range sch.Sections {
		if section == nil {
			continue
		}

		if section.ModuleFormat == nil {
			section.ModuleFormat = sch.ModuleFormat
		}

		if section.SectionFormat == nil {
			section.SectionFormat = sch.SectionFormat
		}

		ret = append(ret, section.Parse(*sch.Globals)...)
	}
	return ret
}

func (sec Section) Parse(globals map[string]any) []string {
	ret := []string{}

	env := map[string]any{
		"Globals": globals,
		"Key":     sec.Title,
		"Val":     "",
	}

	ret = append(ret, parseTemplate(*sec.SectionFormat.Header, env))
	for _, module := range sec.Modules {
		if module == nil {
			continue
		}

		if module.Format == nil {
			module.Format = sec.ModuleFormat
		}

		ret = append(ret, module.Parse(globals))
	}
	ret = append(ret, parseTemplate(*sec.SectionFormat.Footer, env))

	return ret
}

var ansi = regexp.MustCompile("\x1b\\[(\\d+;?)+m")

var tpl = template.New("base").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
	"shell": func(cmd string) string {
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
	},
	"padLeft": func(str string, char rune, width int) string {
		stripped := ansi.ReplaceAllString(str, "")
		l := utf8.RuneCountInString(stripped)
		return str + strings.Repeat(string(char), width-l)
	},
	"padRight": func(str string, char rune, width int) string {
		stripped := ansi.ReplaceAllString(str, "")
		l := utf8.RuneCountInString(stripped)
		return strings.Repeat(string(char), width-l) + str
	},
	"padCenter": func(str string, char rune, width int) string {
		stripped := ansi.ReplaceAllString(str, "")
		sz := utf8.RuneCountInString(stripped)
		w := float64(width - sz)
		left := int(math.Floor(w / 2))
		right := int(math.Ceil(w / 2))
		return strings.Repeat(string(char), left) + str + strings.Repeat(string(char), right)
	},
	"humanSize": func(pow2 bool, sz int) string {
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
	},
})

func parseTemplate(str string, env map[string]any) string {
	t, err := tpl.Parse(str)
	if err != nil {
		fmt.Println(str)
		panic(err)
	}

	var out bytes.Buffer
	err = t.Execute(&out, env)
	if err != nil {
		panic(err)
	}

	return out.String()
}

func (mod *Module) Parse(globals map[string]any) string {
	env := map[string]any{
		"Globals": globals,
	}
	if mod.Locals != nil {
		parseMap(mod.Locals)
	}

	env["Locals"] = *mod.Locals

	if mod.Val == nil {
		*mod.Val = ""
	}

	key := parseTemplate(*mod.Key, env)
	val := parseTemplate(*mod.Val, env)

	env["Key"] = key
	env["Val"] = val

	out := parseTemplate(*mod.Format, env)
	return out
}
