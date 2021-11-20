# repogen 
repogen is a Golang Codegen that generates database query and mutation end-to-end with its model.

## How it works
repogen describe our given tables from our database connection. Therefore, providing the database connection is compulsory.

To define the database connection credentials We could write the connection inside our `.env` file. Another way to define our DB connection is directly using flag `-creds` with DSN URL format.

## Installation

Using go get command  
```
$ go get github.com/sog01/repogen
```

### Go version
This only supports Go 1.16 or higher.

## Getting Started
Once We have successfully installed the repogen We can run it by typing `repogen` following with the mandatory flags.

The easy way to run at the first time is by running this command below :
```
$ repogen -tables <database table separated by commas> -creds <DSN URL>
```
Or We can write this command using go generate.

### Example Generated Files
The repogen will generate two directories that contains models and repositories. We could specify the directory name or packages using flags `-modelDir`,`-modelPackage`, and `-repositoryPackage`. The interfaces of the generated files should look like this :

```
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
```
The implementation is also generated and ready to use inside our project.

## Flags

- `module`: Define go mod name
- `tables`: Define list of tables to generate (comma separated)
- `creds`: Define db credentials with dsn format
- `env`: Define env path that hold creds information
- `envFile`: Define env filename that hold creds information
- `envPrefix`: Define envPrefix that append on creds information
- `destination`: Define destination of generated file
- `modelPackage`: Define model package
- `modelDir`: Define model directory name
- `repositoryPackage`: Define repository package
- `queryOnly`: Only generate the repository code