# storj-mongodb

## Initial Set-up
Make sure your `PATH` includes the `$GOPATH/bin` directory so your commands can be easily used:
```
export PATH=$PATH:$GOPATH/bin
```

Install [github.com/urfave/cli](https://github.com/urfave/cli), by running:
```
$ go get github.com/urfave/cli
```

Install mongo-driver go package, by running:
```
$ go get go.mongodb.org/mongo-driver/mongo
```

## Build ONCE
```
$ go build storj_mongodb.go
```

## Set-up Files
* Create a `db_property` file, with following contents about a MongoDB instance:
    Host Name
    Port Number
    User Name
    Password
* Create a `storj_config.json` file, with Storj network's configuration information in JSON format
* Store both these files in a `config` folder

## Run the command-line tool
* Get help
```
$ ./storj_mongodb.go -h
```
* Check version
```
$ ./storj_mongodb.go -v
```
* Read MongoDB instance property from a desired file
```
$ ./storj_mongodb.go d ./config/db_property
```
* Read MongoDB instance property from a desired file
```
$ ./storj_mongodb.go c ./config/db_property
* Read and parse Storj network's configuration, in JSON format, from a desired file
```
$ ./storj_mongodb.go s ./config/storj_config.json
```