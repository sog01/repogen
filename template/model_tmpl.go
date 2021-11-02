package template

func (tp *TemplateParser) ParseModelTmpl() (string, error) {
	return tp.execTmpl(`
	type {{.Name}} struct {
		{{range .Fields}} {{.GoName}} {{.GoType}} {{.GoTag}}
		{{end}}
	}
 
	type {{.Name}}List []*{{.Name}}
	`)
}
