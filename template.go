package main

import (
	"html/template"
	"path/filepath"
)

func LoadTemplate(filenames ...string) (t *template.Template, err error) {

	for _, filename := range filenames {
		var b []byte
		b, err = Asset(filename)
		if err != nil {
			return
		}

		s := string(b)
		name := filepath.Base(filename)

		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}

		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		_, err = tmpl.Parse(s)
		if err != nil {
			return
		}
	}

	return
}
