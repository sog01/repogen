package parser

import (
	"fmt"
	"strings"
)

type GoResolver struct {
	t *tableDescribe
}

type GoStruct struct {
	Name             string
	IdName           string
	IdType           string
	Fields           []*GoField
	ImportedPackages []string
}

type GoField struct {
	Name        string
	NullType    string
	NullTypeSel string
	Type        string
	Tag         string
}

func (goResolver *GoResolver) ResolveStruct() (*GoStruct, error) {
	goStruct := &GoStruct{
		Name: snakeToCamel(goResolver.t.Name),
	}

	for _, col := range goResolver.t.Columns {
		goField, isId, err := goResolver.ResolveField(col)
		if err != nil {
			return nil, err
		}
		if isId {
			goStruct.IdName = goField.Name
			goStruct.IdType = goField.Type
		}
		goStruct.Fields = append(goStruct.Fields, goField)
	}

	goStruct.ImportedPackages = resolveImportedPkg(goStruct.Fields)

	return goStruct, nil
}

func (goResolver *GoResolver) ResolveField(c *columnDescribe) (*GoField, bool, error) {
	nullable := strings.ToLower(c.Null.String) != "no"
	goType := goResolver.ResolveType(c.Type.String, nullable)
	goNullType, goNullTypeSel := goResolver.resolveNullType(c.Type.String)
	if goType == "unknown" ||
		goNullType == "unknown" {
		return nil, false, fmt.Errorf("unknown '%s' nullable '%v' type",
			c.Type.String,
			nullable)
	}

	goField := &GoField{
		Name:        snakeToCamel(c.Field.String),
		Type:        goType,
		NullType:    goNullType,
		NullTypeSel: goNullTypeSel,
		Tag:         fmt.Sprintf("`db:%s`", `"`+c.Field.String+`"`),
	}

	isId := c.Key.String == "PRI"
	return goField, isId, nil
}

func (goResolver *GoResolver) ResolveType(s string, nullable bool) string {
	s = sanitizeTableType(s)
	if nullable {
		nullType, _ := goResolver.resolveNullType(s)
		return nullType
	}

	switch sanitizeTableType(s) {
	case "bigint":
		return "int64"
	case "int":
		return "int32"
	case "text", "varchar", "enum", "char", "longtext", "mediumblob":
		return "string"
	case "float":
		return "float64"
	case "tinyint":
		return "int8"
	case "datetime", "date", "timestamp":
		return "time.Time"
	case "decimal":
		return "decimal.Decimal"
	default:
		return "unknown"
	}
}

func (goResolver *GoResolver) resolveNullType(s string) (string, string) {
	switch sanitizeTableType(strings.ToLower(s)) {
	case "bigint", "int", "tinyint":
		return "null.Int", "Int64"
	case "text", "varchar", "enum", "char", "longtext", "mediumblob":
		return "null.String", "String"
	case "float":
		return "null.Float", "Float64"
	case "datetime", "date", "timestamp":
		return "null.Time", "Time"
	case "decimal":
		return "decimal.NullDecimal", "NullDecimal"
	default:
		return "unknown", ""
	}
}

func sanitizeTableType(s string) string {
	bracketIndex := strings.Index(s, "(")
	if bracketIndex > -1 {
		return s[:bracketIndex]
	}

	return s
}

func snakeToCamel(s string) string {
	splitted := strings.Split(s, "_")
	for i := range splitted {
		splitted[i] = strings.Title(splitted[i])
	}
	return strings.Title(strings.Join(splitted, ""))
}

func resolveImportedPkg(goFields []*GoField) []string {
	importedMap := make(map[string]struct{})
	for _, field := range goFields {
		if strings.Contains(strings.ToLower(string(field.Type)), "decimal") {
			importedMap["github.com/shopspring/decimal"] = struct{}{}
		} else if strings.Contains(strings.ToLower(string(field.Type)), "null") {
			importedMap["github.com/guregu/null"] = struct{}{}
		} else if strings.Contains(strings.ToLower(string(field.Type)), "time") {
			importedMap["time"] = struct{}{}
		}
	}

	var imported []string
	for importedPkg := range importedMap {
		imported = append(imported, importedPkg)
	}

	return imported
}
