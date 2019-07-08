# storj-mongodb

## Initial Set-up
Make sure your `PATH` includes the `$GOPATH/bin` directory, so that your commands can be easily used [Refer: Install the Go Tools](https://golang.org/doc/install):
```
export PATH=$PATH:$GOPATH/bin
```

Install [github.com/urfave/cli](https://github.com/urfave/cli), by running:
```
$ go get github.com/urfave/cli
```

Install [mongo-driver](https://godoc.org/go.mongodb.org/mongo-driver) go package, by running:
```
$ go get go.mongodb.org/mongo-driver
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
    Database Name

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
```
{ 
    "apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
    "satellite":  "mars.tardigrade.io:7777",
    "bucket":     "my-first-bucket",
	"uploadPath": "foo/bar/baz"
}
```

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
$ ./storj_mongodb.go c ./config/db_property
```

* Read MongoDB instance property from a desired file
```
$ ./storj_mongodb.go d ./config/db_property
```

* Read and parse Storj network's configuration, in JSON format, from a desired file
```
$ ./storj_mongodb.go s ./config/storj_config.json
```