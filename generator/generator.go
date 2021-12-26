package generator

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"

	"github.com/sog01/repogen/parser"
	"github.com/sog01/repogen/template"

	"github.com/jmoiron/sqlx"
)

type Generator struct {
	objParser   *parser.ObjectParser
	fileGens    []*fileGen
	module      string
	destination string
	tables      []string
	opt         *generatorOpt
}

type generatorOpt struct {
	modelPackage      string
	modelDir          string
	repositoryPackage string
	queryOnly         bool
}

type fileGen struct {
	name    string
	tmpl    string
	destDir string
}

func NewGenerator(db *sqlx.DB, module, destination string, tables []string) *Generator {
	return &Generator{
		objParser:   parser.NewTableParser(db),
		module:      module,
		tables:      tables,
		destination: destination,
		opt: &generatorOpt{
			modelPackage:      "model",
			repositoryPackage: "repository",
		},
	}
}

func (gen *Generator) SetModelPackage(modelPackage string) {
	if modelPackage != "" {
		gen.opt.modelPackage = modelPackage
	}
}

func (gen *Generator) SetModelDir(modelDir string) {
	if modelDir != "" {
		gen.opt.modelDir = modelDir
	}
}

func (gen *Generator) SetRepositoryPackage(repositoryPackage string) {
	if repositoryPackage != "" {
		gen.opt.repositoryPackage = repositoryPackage
	}
}

func (gen *Generator) SetQueryOnly(queryOnly bool) {
	gen.opt.queryOnly = queryOnly
}

func (gen *Generator) Generate() error {
	for _, table := range gen.tables {
		obj, err := gen.objParser.Parse(table)
		if err != nil {
			return err
		}

		modelGen, err := gen.genModel(obj)
		if err != nil {
			return err
		}

		repoGen, err := gen.genRepo(obj, gen.resolveModelPath(modelGen.destDir))
		if err != nil {
			return err
		}

		gen.fileGens = append(gen.fileGens, modelGen)
		gen.fileGens = append(gen.fileGens, repoGen...)
	}

	repoHelperGen, err := gen.genRepoArgs()
	if err != nil {
		return err
	}
	gen.fileGens = append(gen.fileGens, repoHelperGen)

	for _, file := range gen.fileGens {
		destDir := gen.destination + "/" + file.destDir
		os.Mkdir(destDir, os.ModePerm)
		writer, err := os.Create(destDir + "/" + file.name)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, file.tmpl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (gen *Generator) resolveModelPath(modelDest string) string {
	destinationPath, err := filepath.Abs(gen.destination)
	if err != nil {
		return modelDest
	}

	splittedDestPath := strings.Split(destinationPath, string(os.PathSeparator))
	n := len(splittedDestPath)
	var modelPath []string
	for i := n - 1; i > 0; i-- {
		dirName := splittedDestPath[i]
		splittedModule := strings.Split(gen.module, "/")
		moduleDirName := splittedModule[len(splittedModule)-1]
		if dirName == moduleDirName {
			modelPath = reverseModelPath(modelPath)
			modelPath = append(modelPath, modelDest)
			return strings.Join(modelPath, "/")
		}
		modelPath = append(modelPath, dirName)
	}

	return modelDest
}

func reverseModelPath(modelPath []string) []string {
	var reversedModelPath []string
	for i := len(modelPath) - 1; i >= 0; i-- {
		reversedModelPath = append(reversedModelPath, modelPath[i])
	}

	return reversedModelPath
}

func (gen *Generator) genModel(obj *parser.Object) (*fileGen, error) {
	tmpl := template.TemplateParser{Object: obj}
	modelTmpl, err := tmpl.ParseModelTmpl()
	if err != nil {
		return nil, err
	}

	var importedPackages []*template.ImportedPackage
	for _, imported := range obj.ImportedPackages {
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name: imported,
		})
	}

	importedTmpl, err := tmpl.ParsePackages(gen.opt.modelPackage, importedPackages)
	if err != nil {
		return nil, err
	}

	modelTmpl = fmt.Sprintf(`%s
	%s`, importedTmpl, modelTmpl)

	formatted, err := format.Source([]byte(modelTmpl))
	if err != nil {
		return nil, err
	}

	destDir := gen.opt.modelDir
	if destDir == "" {
		destDir = gen.opt.modelPackage
	}
	return &fileGen{
		name:    obj.Table + "_gen.go",
		tmpl:    string(formatted),
		destDir: destDir,
	}, nil
}

func (gen *Generator) genRepo(obj *parser.Object, modelPath string) ([]*fileGen, error) {
	repoQueryTmpl, err := gen.genRepoQuery(obj, modelPath)
	if err != nil {
		return nil, err
	}

	if gen.opt.queryOnly {
		return []*fileGen{repoQueryTmpl}, nil
	}

	repoCommandTmpl, err := gen.genRepoCommand(obj, modelPath)
	if err != nil {
		return nil, err
	}

	return []*fileGen{
		repoQueryTmpl,
		repoCommandTmpl,
	}, nil
}

func (gen *Generator) genRepoQuery(obj *parser.Object, modelPath string) (*fileGen, error) {
	tmpl := template.TemplateParser{Object: obj}
	var importedPackages []*template.ImportedPackage
	for _, imported := range repositoryQueryPackages {
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name: imported,
		})
	}
	if gen.opt.repositoryPackage != gen.opt.modelDir {
		tmpl.ModelPackage = obj.LowerName + "model."
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name:  fmt.Sprintf("%s/%s", gen.module, modelPath),
			Alias: obj.LowerName + "model",
		})
	}

	repoTmpl, err := tmpl.ParseRepoQueryImpl()
	if err != nil {
		return nil, err
	}

	importedTmpl, err := tmpl.ParsePackages(gen.opt.repositoryPackage, importedPackages)
	if err != nil {
		return nil, err
	}

	repoTmpl = fmt.Sprintf(`%s
	%s`, importedTmpl, repoTmpl)

	formatted, err := format.Source([]byte(repoTmpl))
	if err != nil {
		return nil, err
	}

	return &fileGen{
		name:    obj.Table + "_repo_query_gen.go",
		tmpl:    string(formatted),
		destDir: gen.opt.repositoryPackage,
	}, nil
}

func (gen *Generator) genRepoCommand(obj *parser.Object, modelPath string) (*fileGen, error) {
	tmpl := template.TemplateParser{Object: obj}
	var importedPackages []*template.ImportedPackage
	for _, imported := range repositoryCommandPackages {
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name: imported,
		})
	}
	if gen.opt.repositoryPackage != gen.opt.modelDir {
		tmpl.ModelPackage = obj.LowerName + "model."
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name:  fmt.Sprintf("%s/%s", gen.module, modelPath),
			Alias: obj.LowerName + "model",
		})
	}

	repoCommandTmpl, err := tmpl.ParseRepoCommandTmpl()
	if err != nil {
		return nil, err
	}

	importedTmpl, err := tmpl.ParsePackages(gen.opt.repositoryPackage, importedPackages)
	if err != nil {
		return nil, err
	}

	repoCommandTmpl = fmt.Sprintf(`%s
	%s`, importedTmpl, repoCommandTmpl)

	formatted, err := format.Source([]byte(repoCommandTmpl))
	if err != nil {
		return nil, err
	}

	return &fileGen{
		name:    obj.Table + "_repo_command_gen.go",
		tmpl:    string(formatted),
		destDir: gen.opt.repositoryPackage,
	}, nil
}

func (gen *Generator) genRepoArgs() (*fileGen, error) {
	tmpl := template.TemplateParser{}
	repoArgs, err := tmpl.ParseRepositoryArgs()
	if err != nil {
		return nil, err
	}

	repoInternalFunc, err := tmpl.ParseInternalFunc()
	if err != nil {
		return nil, err
	}

	var importedPackages []*template.ImportedPackage
	for _, imported := range repositoryArgsPackages {
		importedPackages = append(importedPackages, &template.ImportedPackage{
			Name: imported,
		})
	}

	importedTmpl, err := tmpl.ParsePackages(gen.opt.repositoryPackage, importedPackages)
	if err != nil {
		return nil, err
	}

	repoArgsTmpl := fmt.Sprintf(`%s
	%s
	%s`, importedTmpl, repoArgs, repoInternalFunc)

	formatted, err := format.Source([]byte(repoArgsTmpl))
	if err != nil {
		return nil, err
	}

	return &fileGen{
		name:    "repo_args_gen.go",
		tmpl:    string(formatted),
		destDir: gen.opt.repositoryPackage,
	}, nil
}

var (
	repositoryQueryPackages = []string{
		"context",
		"fmt",
		"strings",
		"errors",
		"github.com/jmoiron/sqlx",
	}

	repositoryCommandPackages = []string{
		"context",
		"github.com/jmoiron/sqlx",
		"strings",
		"database/sql",
		"fmt",
	}

	repositoryArgsPackages = []string{
		"database/sql",
	}
)
