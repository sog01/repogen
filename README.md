# repogen 
repogen is a repository database generator that generates database query and mutation end-to-end with the database model.

## How it works
repogen describe our given tables from our database connection. Therefore, providing the database connection is compulsory.

To define the database connection credentials We could write the connection inside our `.env` file. Another way to define our DB connection is directly using flag `-creds` with DSN URL format.

## Installation

Using go get command  
```go get github.com/sog01/repogen```

### Go version
This only supports Go 1.16 or higher.

## Running repogen
Once We have been successfully intall the repogen We can run by typing `repogen` following the mandatory flags.

The easy way to run at the first time is by run this command below :
```
repogen -tables <database table> -creds <DSN URL>
```
Or We can write this command using go generate.

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