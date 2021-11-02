package template

import (
	"bytes"
	"html/template"

	"github.com/sog01/repogen/parser"
)

type TemplateParser struct {
	Object *parser.Object
}

func (tp *TemplateParser) execTmpl(s string) (string, error) {
	var data struct {
		*parser.Object
		Backtick     string
		OpenBracket  string
		CloseBracket string
	}

	data.Object = tp.Object
	data.Backtick = "`"
	data.OpenBracket = "{"
	data.CloseBracket = "}"
	return execTmpl(s, data)
}

func execTmpl(s string, data interface{}) (string, error) {
	t, err := template.New("").Parse(s)

	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, data)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}
