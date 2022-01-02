package parser

import (
	"database/sql"
	"fmt"
	"html/template"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/jmoiron/sqlx"
)

type ObjectParser struct {
	db        *sqlx.DB
	pluralize *pluralize.Client
}

type Object struct {
	Name                        string
	IdName                      string
	IdDBName                    string
	IdType                      string
	Table                       string
	PrivateName                 string
	LowerName                   string
	ImportedPackages            []string
	Fields                      []*Field
	DBFieldsSeperatedCommas     string
	PlaceholdersSeparatedCommas string
}

type Field struct {
	AutoIncrement     bool
	ObjectName        template.HTML
	ObjectPrivateName template.HTML
	GoName            template.HTML
	GoType            template.HTML
	GoNullType        template.HTML
	GoNullTypeSel     template.HTML
	GoTag             template.HTML
	DBField           template.HTML
}

type tableDescribe struct {
	Name    string
	Columns []*columnDescribe
}

type columnDescribe struct {
	Field   sql.NullString `db:"Field"`
	Type    sql.NullString `db:"Type"`
	Null    sql.NullString `db:"Null"`
	Key     sql.NullString `db:"Key"`
	Default sql.NullString `db:"Default"`
	Extra   sql.NullString `db:"Extra"`
}

func NewTableParser(db *sqlx.DB, pluralize *pluralize.Client) *ObjectParser {
	return &ObjectParser{
		db:        db,
		pluralize: pluralize,
	}
}

func (tp *ObjectParser) Parse(table string) (*Object, error) {
	tableDescribe, err := tp.parseTable(table)
	if err != nil {
		return nil, err
	}

	goResolver := GoResolver{tableDescribe}
	goStruct, err := goResolver.ResolveStruct()
	if err != nil {
		return nil, err
	}

	obj := &Object{
		Name:             goStruct.Name,
		IdName:           strings.ToLower(goStruct.IdName),
		IdType:           goStruct.IdType,
		Table:            table,
		PrivateName:      strings.ToLower(goStruct.Name[0:1]) + goStruct.Name[1:],
		LowerName:        strings.ToLower(goStruct.Name),
		ImportedPackages: goStruct.ImportedPackages,
	}

	var (
		dbFields     []string
		placeholders []string
	)
	for index, goField := range goStruct.Fields {
		column := tableDescribe.Columns[index]
		if column.Key.String == "PRI" {
			obj.IdDBName = column.Field.String
		}

		// make the name as singular if the tables name is a plurar
		obj.Name = tp.pluralize.Singular(obj.Name)
		obj.PrivateName = tp.pluralize.Singular(obj.PrivateName)

		autoIncrement := column.Extra.String == "auto_increment"
		obj.Fields = append(obj.Fields, &Field{
			AutoIncrement:     autoIncrement,
			ObjectName:        template.HTML(obj.Name),
			ObjectPrivateName: template.HTML(obj.PrivateName),
			GoName:            template.HTML(goField.Name),
			GoType:            template.HTML(goField.Type),
			GoNullType:        template.HTML(goField.NullType),
			GoNullTypeSel:     template.HTML(goField.NullTypeSel),
			GoTag:             template.HTML(goField.Tag),
			DBField:           template.HTML(column.Field.String),
		})
		if !autoIncrement {
			dbFields = append(dbFields, column.Field.String)
			placeholders = append(placeholders, "?")
		}
	}
	obj.DBFieldsSeperatedCommas = strings.Join(dbFields, `,
	`)
	obj.PlaceholdersSeparatedCommas = strings.Join(placeholders, `,
	`)
	return obj, nil
}

func (tp *ObjectParser) parseTable(table string) (*tableDescribe, error) {
	columnDescribes := []*columnDescribe{}
	err := tp.db.Select(
		&columnDescribes,
		fmt.Sprintf("DESCRIBE `%s`", table))
	if err != nil {
		return nil, err
	}

	return &tableDescribe{table, columnDescribes}, nil
}
