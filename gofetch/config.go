package gofetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"text/template"

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
	Each   []string
}

type Section struct {
	Title         string
	Modules       []*Module
	ModuleFormat  *string
	SectionFormat *SecFormat
}

func parseMap(obj *map[string]any, env *map[string]any) {
	if env == nil {
		env = obj
	}
	for k, v := range *obj {
		if r := reflect.TypeOf(v); r.Kind() == reflect.Map {
			m := v.(map[string]any)
			if m["Cast"] == nil {
				continue
			}

			if m["Debug"] != nil {
				fmt.Println(env)
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
			case "Map":
				q := map[string]any{}
				x = parseTemplate(x, *env)
				err = json.Unmarshal([]byte(x), &q)
				(*obj)[k] = q
			case "List":
				q := []map[string]any{}
				x = parseTemplate(x, *env)
				err = json.Unmarshal([]byte(x), &q)
				if err != nil {
					qq := []any{}
					old_err := err
					err = json.Unmarshal([]byte(x), &qq)
					if err != nil {
						fmt.Println(x)
						panic(old_err)
					}
					(*obj)[k] = qq
				} else {
					(*obj)[k] = q
				}
			default:
				panic(fmt.Sprintf("cast not implemented: %s", m["Cast"]))
			}
			if err != nil {
				panic(err)
			}
			if m["Debug"] != nil {
				fmt.Printf("\x1b[1m%s:\x1b[0m\n\t<==%s\n\t==>%s\n\n", k, x, (*obj)[k])
			}
			continue
		}
		if r := reflect.TypeOf(v); r.Kind() != reflect.String {
			continue
		}
		(*obj)[k] = parseTemplate(fmt.Sprint(v), *env)
	}
}

func (sch Schema) Parse() []string {
	if sch.Globals != nil {
		m := map[string]any{
			"Globals": *sch.Globals,
		}
		parseMap(sch.Globals, &m)
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

		ret = append(ret, module.Parse(globals)...)
	}
	ret = append(ret, parseTemplate(*sec.SectionFormat.Footer, env))

	return ret
}

var ansi = regexp.MustCompile("\x1b\\[(\\d+;?)+m")

var tpl = template.New("base").Funcs(sprig.FuncMap()).Funcs(template.FuncMap{
	"shell":     tmpl_Shell,
	"padLeft":   tmpl_PadLeft,
	"padRight":  tmpl_PadRight,
	"padCenter": tmpl_PadCenter,
	"humanSize": tmpl_HumanSize,
	"yank":      tmpl_Yank,
	"atMap":     tmpl_AtMap,
	"at":        tmpl_At,
	"key":       tmpl_Key,
	"tee":       func(q ...any) []any { fmt.Println(q...); return q },
	"teeMap":    func(q map[string]any) map[string]any { fmt.Println(q); return q },
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
		fmt.Println(str)
		panic(err)
	}

	return out.String()
}

func (mod *Module) Parse(globals map[string]any) []string {
	env := map[string]any{
		"Globals": globals,
	}

	env["Locals"] = *mod.Locals

	if mod.Locals != nil {
		parseMap(mod.Locals, &env)
	}

	if mod.Val == nil {
		*mod.Val = ""
	}

	ret := []string{}
	run := func() {
		key := parseTemplate(*mod.Key, env)
		val := parseTemplate(*mod.Val, env)

		env["Key"] = key
		env["Val"] = val

		ret = append(ret, parseTemplate(*mod.Format, env))
	}

	if mod.Each != nil {
		var loop any
		loop = env
		for _, k := range mod.Each {
			loop = (loop.(map[string]any))[k]
		}

		try, ok := loop.([]any)
		if !ok {
			try2 := loop.([]map[string]any)

			for i := range try2 {
				env["Idx"] = i
				run()
			}
		} else {
			for i := range try {
				env["Idx"] = i
				run()
			}
		}

	} else {
		run()
	}
	return ret
}
