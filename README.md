# storj-mongodb

# Initial Set-up
* Create a folder "github.com" within $GOPATH/src
* Create a folder "urfave" within $GOPATH/src/github.com
* Clone "github.com/urfave/cli" within $GOPATH/src/github.com/urfave

# Build ONCE
```
$ go build storj_mongodb.go
```

# Run the command-line tool
```
$ ./storj_mongodb.go -h
$ ./storj_mongodb.go -v
$ ./storj_mongodb.go d
$ ./storj_mongodb.go s
```