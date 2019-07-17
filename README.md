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

Install [storj-uplink](https://godoc.org/storj.io/storj/lib/uplink) go package, by running:
```
$ go get storj.io/storj/lib/uplink
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
	"uploadPath": "foo/bar/baz",
    "encryptionpassphrase": "test"
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

* Read BSON data from desired MongoDB instance to given Storj network bucket
```
$ ./storj_mongodb.go c ./config/db_property ./config/storj_config.json
```

* Unit Test: MongoDB instance's property read from a desired file, establish connection and read all collections
```
$ ./storj_mongodb.go d ./config/db_property
```

* Unit Test: Storj network's configuration read and parsed, in JSON format, from a desired file, establish connection and send a sample data to a bucket
```
$ ./storj_mongodb.go s ./config/storj_config.json
```