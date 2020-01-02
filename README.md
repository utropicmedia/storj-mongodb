# storj-mongodb
### Developed using libuplink version : v0.27.1

## Install and configure- Go
* Install Go for your platform by following the instructions in given link
[Refer: Installing Go](https://golang.org/doc/install#install)

* Make sure your `PATH` includes the `$GOPATH/bin` directory, so that your commands can be easily used:
```
export PATH=$PATH:$GOPATH/bin
```

## Setting up Storj-Mongodb project

* Put the utropicmedia folder in ***`go/src`*** folder in your home directoy.

* Put the storj-mongodb folder in ***`go/src`*** folder in your home directory.

* Now open `terminal`, navigate to the `storj-mongodb` project folder and download following dependencies one by one required by the project:

```
$ go get -u github.com/urfave/cli
$ go get -u go.mongodb.org/mongo-driver
$ go get -u storj.io/storj/lib/uplink
$ go get -u ./...
```

## Set-up Files
* Create a `db_property.json` file, with following contents about a MongoDB instance:
    * hostName :- Host Name connect to MongoDB
    * port :- Port connect to MongoDB
    * username :- User Name of MongoDB
    * password :- Password of MongoDB
    * database :- MongoDB Database Name

```json
    { 
        "hostname": "mongodbHostName",
        "port":     "27017",
        "username": "username",
        "password": "password",
        "database": "mongoDatabaseName"
    }
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
    * apiKey :- API key created in Storj satellite gui
    * satelliteURL :- Storj Satellite URL
    * encryptionPassphrase :- Storj Encryption Passphrase.
    * bucketName :- Split file into given size before uploading.
    * uploadPath :- Path on Storj Bucket to store data (optional) or "/"
    * serializedScope:- Serialized Scope Key shared while uploading data used to access bucket without API key
    * disallowReads:- Set true to create serialized scope key with restricted read access
    * disallowWrites:- Set true to create serialized scope key with restricted write access
    * disallowDeletes:- Set true to create serialized scope key with restricted delete access

```json
    { 
        "apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
        "satellite":  "us-central-1.tardigrade.io:7777",
        "bucket":     "change-me-to-desired-bucket-name",
        "uploadPath": "optionalpath/requiredfilename",
        "encryptionpassphrase": "you'll never guess this",
        "serializedScope": "change-me-to-the-api-key-created-in-encryption-access-apiKey",
        "disallowReads": "true/false-to-disallow-reads",
        "disallowWrites": "true/false-to-disallow-writes",
        "disallowDeletes": "true/false-to-disallow-deletes"
    }
```

* Store both these files in a `config` folder.  Filename command-line arguments are optional.  defualt locations are used.


## Build ONCE
```
$ go build storj_mongodb.go
```

## Run the command-line tool

**NOTE**: The following commands operate in a Linux system

* Get help
```
    $ ./storj_mongodb -h
```

* Check version
```
    $ ./storj_mongodb -v
```

* Read BSON data from desired MongoDB instance and upload it to given Storj network bucket using Serialized Scope Key.  [note: filename arguments are optional.  default locations are used.]
```
    $ ./storj_mongodb store ./config/db_property.json ./config/storj_config.json  
```

* Read BSON data from desired MongoDB instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates an unrestricted shareable Serialized Scope Key.  [note: filename arguments are optional. default locations are used.]
```
    $ ./storj_mongodb store ./config/db_property.json ./config/storj_config.json key
```

* Read BSON data from desired MongoDB instance and upload it to given Storj network bucket API key and EncryptionPassPhrase from storj_config.json and creates a restricted shareable Serialized Scope Key.  [note: filename arguments are optional. default locations are used. `restrict` can only be used with `key`]
```
    $ ./storj_mongodb store ./config/db_property.json ./config/storj_config.json key restrict
```

* Read BSON data in `debug` mode from desired MongoDB instance and upload it to given Storj network bucket.  [note: filename arguments are optional.  default locations are used. Make sure `debug` folder already exist in project folder.]
```
    $ ./storj_mongodb store debug ./config/db_property.json ./config/storj_config.json  
```

* Read MongoDB instance property from a desired JSON file and display all its collections' data
```
    $ ./storj_mongodb parse   
```

* Read MongoDB instance property in `debug` mode from a desired JSON file and display all its collections' data
```
    $ ./storj_mongodb.go parse debug 
```

* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object
```
    $ ./storj_mongodb.go test 
```
* Read and parse Storj network's configuration, in JSON format, from a desired file and upload a sample object in `debug` mode
```
    $ ./storj_mongodb.go test debug 
```
