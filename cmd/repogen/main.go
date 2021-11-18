package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sog01/repogen/generator"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func main() {
	module := flag.String("module", "", "define go mod name")
	tables := flag.String("tables", "", "comma separated list of tables to generate")
	dbCreds := flag.String("creds", "", "define db credentials with dsn format")
	dbEnv := flag.String("env", "", "define env that hold creds information")
	dbEnvFile := flag.String("envFile", "", "define env that hold creds information")
	dbEnvPrefix := flag.String("envPrefix", "", "define envPrefix that append on creds information")
	destination := flag.String("destination", "", "define destination")
	modelPackage := flag.String("modelPackage", "", "define model package")
	modelDir := flag.String("modelDir", "", "define model directory name")
	repositoryPackage := flag.String("repositoryPackage", "", "define repository package")
	ignoreError := flag.Bool("ignoreError", false, "ignore the error that occurs which probably happen in CI / CD")
	flag.Parse()

	err := generate(*module,
		*tables,
		*dbCreds,
		*dbEnv,
		*dbEnvFile,
		*dbEnvPrefix,
		*destination,
		*modelPackage,
		*modelDir,
		*repositoryPackage)
	if err != nil && !*ignoreError {
		log.Fatal(err)
	}
}

func generate(module,
	tables,
	dbCreds,
	dbEnv,
	dbEnvFile,
	dbEnvPrefix,
	destination,
	modelPackage,
	modelDir,
	repositoryPackage string) error {
	if len(tables) == 0 {
		log.Fatal("empty tables")
	}

	if module == "" {
		log.Fatal("empty module")
	}

	if destination == "" {
		destination = "./"
	}

	creds := dbCreds
	if creds == "" {
		var err error
		if dbEnvFile != "" {
			creds, err = readCredsFromEnvFile(dbEnvFile, dbEnvPrefix, module)
		} else {
			creds, err = readCredsFromEnv(dbEnv, dbEnvPrefix)
		}

		if err != nil {
			return err
		}
	}

	db, err := sqlx.Open("mysql", creds)
	if err != nil {
		return errors.New("unable to connect to db")
	}
	gen := generator.NewGenerator(db, module, destination, strings.Split(tables, ","))
	gen.SetModelPackage(modelPackage)
	gen.SetModelDir(modelDir)
	gen.SetRepositoryPackage(repositoryPackage)
	if err := gen.Generate(); err != nil {
		return fmt.Errorf("unable to generate repository: %v", err)
	}

	return nil
}

func readCredsFromEnv(envPath, prefixenv string) (string, error) {
	if err := godotenv.Load(envPath); err != nil {
		return "", errors.New("unable to load env file")
	}

	prefix := prefixenv
	if prefix != "" {
		prefix = prefix + "."
	}

	dsn := fmt.Sprintf("%s:%s@(%s:%s)/%s", os.Getenv(prefix+"REPOGEN.DB.USERNAME"),
		os.Getenv(prefix+"REPOGEN.DB.PASSWORD"),
		os.Getenv(prefix+"REPOGEN.DB.HOST"),
		os.Getenv(prefix+"REPOGEN.DB.PORT"),
		os.Getenv(prefix+"REPOGEN.DB.DATABASE"))

	return dsn, nil
}

func readCredsFromEnvFile(envFile, prefixEnv, module string) (string, error) {
	rootDirPath, err := findRootDirPath(module)
	if err != nil {
		return "", nil
	}

	var envPath string
	filepath.Walk(rootDirPath, func(path string, info fs.FileInfo, err error) error {
		if len(path) < 4 {
			return nil
		}
		fileName := path[len(path)-4:]
		if fileName == envFile {
			envPath = path
		}
		return nil
	})

	return readCredsFromEnv(envPath, prefixEnv)
}

func findRootDirPath(module string) (string, error) {
	currDirPath, err := os.Getwd()
	if err != nil {
		return "", errors.New("unable to get working directory")
	}

	splittedCurrDir := strings.Split(currDirPath, "/")
	splittedModule := strings.Split(module, "/")
	moduleDir := splittedModule[len(splittedModule)-1]
	var rootDirPath []string
	for _, dir := range splittedCurrDir {
		rootDirPath = append(rootDirPath, dir)
		if dir == moduleDir {
			break
		}
	}

	return strings.Join(rootDirPath, "/"), nil
}
