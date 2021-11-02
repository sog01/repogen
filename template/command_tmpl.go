package template

func (tp *TemplateParser) ParseRepoCommandTmpl() (string, error) {
	return tp.execTmpl(`
	type Repository{{.Name}}Command interface {
		Insert{{.Name}}List(ctx context.Context, {{.PrivateName}}List {{.LowerName}}model.{{.Name}}List) (*InsertResult, error)
		Insert{{.Name}}(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}) (*InsertResult, error)
		Update{{.Name}}ByFilter(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}, filter Filter, updatedFields ...{{.Name}}Field) error
		Update{{.Name}}(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}, {{.IdName}} {{.IdType}}, updatedFields ...{{.Name}}Field) error
		Delete{{.Name}}List(ctx context.Context, filter Filter) error
		Delete{{.Name}}(ctx context.Context, {{.IdName}} {{.IdType}}) error
	}

	type Repository{{.Name}}CommandImpl struct {
		db   *sqlx.DB
		tx   *sqlx.Tx
	}

	func(repo *Repository{{.Name}}CommandImpl) Insert{{.Name}}List(ctx context.Context, {{.PrivateName}}List {{.LowerName}}model.{{.Name}}List) (*InsertResult, error) {
		command := {{.Backtick}}INSERT INTO {{.Table}} ({{.DBFieldsSeperatedCommas}}) VALUES
		{{.Backtick}}

		var (
			placeholders []string
			args   []interface{}
		)
		for _, {{.PrivateName}} := range {{.PrivateName}}List {
			placeholders = append(placeholders, {{.Backtick}}{{.PlaceholdersSeparatedCommas}}{{.Backtick}})
			args = append(args, {{range .Fields}}{{if .AutoIncrement}}
				{{else}}{{.ObjectPrivateName}}.{{.GoName}},
				{{end}}{{end}}
			)
		}
		command += strings.Join(placeholders, ",")
		
		sqlResult, err := repo.exec(ctx, command, args)
		if err != nil {
			return nil, err
		}
		
		return &InsertResult{Result: sqlResult}, nil
	}

	func(repo *Repository{{.Name}}CommandImpl) Insert{{.Name}}(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}) (*InsertResult, error) {
		return repo.Insert{{.Name}}List(ctx, {{.LowerName}}model.{{.Name}}List{{.OpenBracket}}{{.PrivateName}}{{.CloseBracket}})
	}

	func(repo *Repository{{.Name}}CommandImpl) Update{{.Name}}ByFilter(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}, filter Filter, updatedFields ...{{.Name}}Field) error {
		updatedFieldQuery, values := buildUpdateFields{{.Name}}Query(updatedFields, {{.PrivateName}})
		command := fmt.Sprintf({{.Backtick}}UPDATE {{.Table}} 
			SET %s 
		WHERE %s
		{{.Backtick}}, strings.Join(updatedFieldQuery, ","), filter.Query())
		values = append(values, filter.Values()...)
		_, err := repo.exec(ctx, command, values)
		return err
	}

	func(repo *Repository{{.Name}}CommandImpl) Update{{.Name}}(ctx context.Context, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}, {{.IdName}} {{.IdType}}, updatedFields ...{{.Name}}Field) error {
		updatedFieldQuery, values := buildUpdateFields{{.Name}}Query(updatedFields, {{.PrivateName}})
		command := fmt.Sprintf({{.Backtick}}UPDATE {{.Table}} 
			SET %s 
		WHERE {{.IdDBName}} = ?
		{{.Backtick}}, strings.Join(updatedFieldQuery, ","))
		values = append(values, {{.IdName}})
		_, err := repo.exec(ctx, command, values)
		return err
	}

	func(repo *Repository{{.Name}}CommandImpl) Delete{{.Name}}List(ctx context.Context, filter Filter) error {
		command := "DELETE FROM {{.Table}} WHERE "+filter.Query()
		_, err := repo.exec(ctx, command, filter.Values())
		return err
	}

	func(repo *Repository{{.Name}}CommandImpl) Delete{{.Name}}(ctx context.Context, {{.IdName}} {{.IdType}}) error {
		command := "DELETE FROM {{.Table}} WHERE {{.IdDBName}} = ?"
		_, err := repo.exec(ctx, command, []interface{{.OpenBracket}}{{.CloseBracket}}{{.OpenBracket}}{{.IdName}}{{.CloseBracket}})
		return err
	}

	func NewRepo{{.Name}}Command(db *sqlx.DB) Repository{{.Name}}Command {
		return &Repository{{.Name}}CommandImpl{
			db: db,
		}
	}

	func NewRepo{{.Name}}CommandFromTx(tx *sqlx.Tx) Repository{{.Name}}Command {
		return &Repository{{.Name}}CommandImpl{
			tx: tx,
		}
	}

	func(repo *Repository{{.Name}}CommandImpl) exec(ctx context.Context, command string, args []interface{}) (sql.Result, error) {
		var (
			stmt *sqlx.Stmt
			err  error
		)
		if repo.tx != nil {
			stmt, err = repo.tx.PreparexContext(ctx, command)
		} else {
			stmt, err = repo.db.PreparexContext(ctx, command)
		}
	
		if err != nil {
			return nil, err
		}
	
		return stmt.ExecContext(ctx, args...)
	}

	func buildUpdateFields{{.Name}}Query(updatedFields {{.Name}}FieldList, {{.PrivateName}} *{{.LowerName}}model.{{.Name}}) ([]string, []interface{}) {
		var (
			updatedFieldsQuery []string
			args        []interface{}
		)

		for _, field := range updatedFields {
			switch field {
			{{range .Fields}} case "{{.DBField}}":
				updatedFieldsQuery = append(updatedFieldsQuery, "{{.DBField}} = ?")
				args = append(args, {{.ObjectPrivateName}}.{{.GoName}})
			{{end}}}
		}


		return updatedFieldsQuery, args
	}
	`)
}
