package mutators

import (
	"bytes"
	"html/template"

	"github.com/dadrus/heimdall/internal/pipeline/subject"
)

type Template string

func (t Template) Render(sub *subject.Subject) (string, error) {
	tmpl, err := template.New("Subject").Parse(string(t))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, sub)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}