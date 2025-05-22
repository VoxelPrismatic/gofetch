package gofetch

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os/exec"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

type Schema struct {
	Sections      []*Section
	Globals       *map[string]any
	ModuleFormat  *string
	SectionFormat *string
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
	SectionFormat *string
}

func (sch Schema) Parse() []string {
	if sch.Globals != nil {
		for k, v := range *sch.Globals {
			if reflect.TypeOf(v).Kind() != reflect.String {
				continue
			}
			(*sch.Globals)[k] = parseTemplate(fmt.Sprint(v), nil)
		}
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
	for _, module := range sec.Modules {
		if module == nil {
			continue
		}

		if module.Format == nil {
			module.Format = sec.ModuleFormat
		}

		ret = append(ret, module.Parse(globals))
	}

	return ret
}

var tpl = template.New("base").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
	"shell": func(cmd string) string {
		call := exec.Command("sh", cmd)
		reader, err := call.StdoutPipe()
		if err != nil {
			panic(err)
		}
		err = call.Run()
		if err != nil {
			panic(err)
		}
		stdout, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		return string(stdout)
	},
	"padLeft": func(str string, char rune, width int) string {
		return str + strings.Repeat(string(char), width-len(str))
	},
	"padRight": func(str string, char rune, width int) string {
		return strings.Repeat(string(char), width-len(str)) + str
	},
	"padCenter": func(str string, char rune, width int) string {
		sz := float64(len(str))
		w := float64(width) - sz
		left := int(math.Floor(w / 2))
		right := int(math.Ceil(w / 2))
		return strings.Repeat(string(char), left) + str + strings.Repeat(string(char), right)
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
		for k, v := range *mod.Locals {
			if reflect.TypeOf(v).Kind() != reflect.String {
				continue
			}
			(*mod.Locals)[k] = parseTemplate(fmt.Sprint(v), nil)
		}
	}

	env["Locals"] = *mod.Locals

	key := parseTemplate(*mod.Key, env)
	val := parseTemplate(*mod.Val, env)

	env["Key"] = key
	env["Val"] = val

	out := parseTemplate(*mod.Format, env)
	return out
}
