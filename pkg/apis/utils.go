package apis

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/chirino/graphql/schema"
	"github.com/pkg/errors"
)

func sanitizeName(id string) string {
	// valid ids have match this regex: `/^[_a-zA-Z][_a-zA-Z0-9]*$/`
	if id == "" {
		return id
	}
	buf := []byte(id)
	c := buf[0]
	if !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || c == '_') {
		buf[0] = '_'
	}
	for i := 1; i < len(buf); i++ {
		c = buf[i]
		if !(('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9') || c == '_') {
			buf[i] = '_'
		}
	}
	return string(buf)
}

func makeUnique(existing map[string]bool, name string) string {
	cur := name
	for i := 1; existing[cur]; i++ {
		cur = fmt.Sprintf("%s%d", name, i)
	}
	existing[cur] = true
	return cur
}

func requiredType(qlType schema.Type, required bool) (t schema.Type) {
	if required {
		return &schema.NonNull{OfType: qlType}
	}
	return qlType
}

func capitalizeFirstLetter(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[0:1]) + name[1:]
}

func description(desc string) string {
	if desc == "" {
		return ""
	}
	if !strings.HasSuffix(desc, "\n") {
		desc += "\n"
	}
	desc = "\n" + `"""` + "\n" + desc + `"""` + "\n"
	return desc
}

func renderTemplate(variables interface{}, templateText string) (string, error) {
	buf := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(templateText)
	if err != nil {
		return "", errors.WithStack(err)
	}
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		return "", errors.WithStack(err)
	}
	result := buf.String()
	return result, nil
}

func desc(text string) schema.Description {
	return schema.NewDescription(strings.TrimSpace(text))
}
