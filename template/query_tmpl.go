package template

func (tp *TemplateParser) ParseRepoQueryImpl() (string, error) {
	return tp.execTmpl(`
	type Repository{{.Name}}Query interface {
		Select{{.Name}}(fields ...{{.Name}}Field) Repository{{.Name}}Query
		Exclude{{.Name}}(excludedFields ...{{.Name}}Field) Repository{{.Name}}Query
		Filter{{.Name}}(filter Filter) Repository{{.Name}}Query
		Pagination{{.Name}}(pagination Pagination) Repository{{.Name}}Query
		OrderBy{{.Name}}(orderBy []Order) Repository{{.Name}}Query
		Get{{.Name}}Count(ctx context.Context) (int, error)
		Get{{.Name}}(ctx context.Context)  (*{{.ModelPackage}}{{.Name}}, error)
		Get{{.Name}}List(ctx context.Context)  ({{.ModelPackage}}{{.Name}}List, error)
	}

	type Repository{{.Name}}QueryImpl struct {
		db   *sqlx.DB
		query string
		filter      Filter
		orderBy     []Order
		pagination  Pagination
		fields      {{.Name}}FieldList
	}

	func (repo *Repository{{.Name}}QueryImpl) Select{{.Name}}(fields ...{{.Name}}Field) Repository{{.Name}}Query {
		return &Repository{{.Name}}QueryImpl{
			db:         repo.db,
			filter:     repo.filter,
			orderBy:    repo.orderBy,
			pagination: repo.pagination,
			fields:     fields,
		}
	}

	func (repo *Repository{{.Name}}QueryImpl) Exclude{{.Name}}(excludedFields ...{{.Name}}Field) Repository{{.Name}}Query {
		selectedFieldsStr := excludeFields({{.Name}}FieldList(excludedFields).toString(), 
			{{.Name}}SelectFields{}.All().toString())

		var selectedFields []{{.Name}}Field
		for _, sel := range selectedFieldsStr {
			selectedFields = append(selectedFields, {{.Name}}Field(sel))
		}

		return &Repository{{.Name}}QueryImpl{
			db:         repo.db,
			filter:     repo.filter,
			orderBy:    repo.orderBy,
			pagination: repo.pagination,
			fields:     selectedFields,
		}
	}

	func (repo *Repository{{.Name}}QueryImpl) Filter{{.Name}}(filter Filter) Repository{{.Name}}Query {
		return &Repository{{.Name}}QueryImpl{
			db:         repo.db,
			filter:     filter,
			orderBy:    repo.orderBy,
			pagination: repo.pagination,
			fields:     repo.fields,
		}
	}

	func (repo *Repository{{.Name}}QueryImpl) Pagination{{.Name}}(pagination Pagination) Repository{{.Name}}Query {
		return &Repository{{.Name}}QueryImpl{
			db:         repo.db,
			filter:     repo.filter,
			orderBy:    repo.orderBy,
			pagination: pagination,
			fields:     repo.fields,
		}
	}

	func (repo *Repository{{.Name}}QueryImpl) OrderBy{{.Name}}(orderBy []Order) Repository{{.Name}}Query {
		return &Repository{{.Name}}QueryImpl{
			db:         repo.db,
			filter:     repo.filter,
			orderBy:    orderBy,
			pagination: repo.pagination,
			fields:     repo.fields,
		}
	}

	func (repo *Repository{{.Name}}QueryImpl) Get{{.Name}}List(ctx context.Context)  ({{.ModelPackage}}{{.Name}}List, error) {
		var (
			{{.PrivateName}}List {{.ModelPackage}}{{.Name}}List
			values []interface{}
		)

		if len(repo.fields) == 0 {
			repo.fields = {{.Name}}SelectFields{}.All()
		}

		query := fmt.Sprintf("SELECT %s FROM {{.Table}}", strings.Join(repo.fields.toString(), ","))
		if repo.filter != nil {
			query += " WHERE "+repo.filter.Query()
			values = append(values, repo.filter.Values()...)
		}

		if len(repo.orderBy) > 0 {
			var orderStr []string
			for _, order := range repo.orderBy {
				orderStr = append(orderStr, order.Value()+" "+order.Direction())
			}
			query += fmt.Sprintf(" ORDER BY %s", strings.Join(orderStr, ","))
		}

		if repo.pagination != nil {
			offset := (repo.pagination.GetPage() - 1) * repo.pagination.GetSize()
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", repo.pagination.GetSize(), offset)
		}

		err := repo.db.SelectContext(ctx, &{{.PrivateName}}List, query, values...)
		if err != nil {
			return nil, err
		}
		return {{.PrivateName}}List, nil
	}

	func (repo *Repository{{.Name}}QueryImpl) Get{{.Name}}Count(ctx context.Context) (int, error) {
		var values []interface{}
		query := fmt.Sprintf("SELECT count(1) FROM {{.Table}}")
		if repo.filter != nil {
			query += " WHERE "+repo.filter.Query()
			values = append(values, repo.filter.Values()...)
		}

		var count int
		err := repo.db.QueryRowContext(ctx, query, values...).Scan(&count)
		return count, err
	}

	func (repo *Repository{{.Name}}QueryImpl) Get{{.Name}}(ctx context.Context)  (*{{.ModelPackage}}{{.Name}}, error) {
		{{.PrivateName}}List, err := repo.Get{{.Name}}List(ctx)
		if err != nil {
			return nil, err
		}

		if len({{.PrivateName}}List) == 0 {
			return nil, errors.New("{{.LowerName}} not found")
		}

		return {{.PrivateName}}List[0], nil
	}

	func NewRepo{{.Name}}Query(db *sqlx.DB) Repository{{.Name}}Query {
		return &Repository{{.Name}}QueryImpl{
			db: db,
		}
	}
	` + templateFields +
		templateFilter +
		templateOrder)
}

func (tp *TemplateParser) ParseRepositoryArgs() (string, error) {
	return tp.execTmpl(`
		type Filter interface {
			Query()   string
			Values()   []interface{}
		}

		type Pagination interface {
			GetPage() int
			GetSize() int
		}

		type Order interface {
			Value()     string
			Direction() string
		}

		type PaginationData struct {
			Page int
			Size int
		}

		func (p PaginationData) GetPage() int {
			return p.Page
		}

		func (p PaginationData) GetSize() int {
			return p.Size
		}

		type InsertResult struct {
			sql.Result
		}
		`)
}

func (tp *TemplateParser) ParseInternalFunc() (string, error) {
	return tp.execTmpl(`
	func excludeFields(excludedFields, allFields []string) []string {
		var selectedFields []string
			for _, field := range allFields {
				var isExField bool
				for _, exField := range excludedFields {
					if field == exField {
						isExField = true
						break
					}
				}
				
				if isExField {
					continue
				}
				selectedFields = append(selectedFields, field)
			}
			return selectedFields
	    }
	`)
}

var (
	templateFields = `
	type {{.Name}}Field string
	type {{.Name}}FieldList []{{.Name}}Field
	func (fieldList {{.Name}}FieldList) toString() []string{
		var fieldsStr []string
		for _, field := range fieldList {
			fieldsStr = append(fieldsStr, string(field))
		}
		return fieldsStr
	}


	type {{.Name}}SelectFields struct {	
	}
	{{range .Fields}}func({{.ObjectName}}SelectFields){{.GoName}}() {{.ObjectName}}Field {
		return {{.ObjectName}}Field("{{.DBField}}")
	}
	{{end}}
	func({{.Name}}SelectFields) All() {{.Name}}FieldList{
		return []{{.Name}}Field{
			{{range .Fields}} {{.ObjectName}}Field("{{.DBField}}"),
			{{end}}}
	}

	func New{{.Name}}SelectFields() {{.Name}}SelectFields {
		return {{.Name}}SelectFields{}
	}
	`
	templateFilter = `type {{.Name}}Filter struct {
		operator string
		query  []string
		values []interface{}
	}

	func New{{.Name}}Filter(operator string) {{.Name}}Filter {
		if operator == "" {
			operator = "AND"
		}
		return {{.Name}}Filter{
			operator: operator,
		}
	}

	{{range .Fields}} func(f {{.ObjectName}}Filter) SetFilterBy{{.GoName}}(value interface{}, operator string) {{.ObjectName}}Filter {
		query := "{{.DBField}} " + operator + " (?)"
		var values []interface{}
		switch strings.ToUpper(operator) {
		case "IN":
			query, values, _ = sqlx.In(query, value)
		default:
			values = append(values, value)
		}
		return {{.ObjectName}}Filter {
			operator: f.operator,
			query:  append(f.query, query),
			values: append(f.values, values...),
		}
	}
	{{end}}

	func (f {{.Name}}Filter) Query() string {
		return strings.Join(f.query, " "+f.operator+" ")
	}

	func (f {{.Name}}Filter) Values() []interface{} {
		return f.values
	}
	`

	templateOrder = `
	{{range .Fields}}type {{.ObjectName}}{{.GoName}}Order struct {
		direction string
	}
	func(o {{.ObjectName}}{{.GoName}}Order)SetDirection(direction string) {{.ObjectName}}{{.GoName}}Order {
		return {{.ObjectName}}{{.GoName}}Order {
			direction: direction,
		}
	}
	func(o {{.ObjectName}}{{.GoName}}Order)Value() string {
		return "{{.DBField}}"
	}
	func(o {{.ObjectName}}{{.GoName}}Order)Direction() string {
		return o.direction
	}
	func New{{.ObjectName}}{{.GoName}}Order() {{.ObjectName}}{{.GoName}}Order {
		return {{.ObjectName}}{{.GoName}}Order{}
	}
	{{end}}
	`
)
