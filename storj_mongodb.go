/*
 * package to read MongoDB collections into a Storj bucket
 *
 * v 1.0.5
 * segregated codes for MongoDB and Storj into separate packages,
 * and called them here to realize the combined interface
 *
 * v 1.0.4
 * reads MongoDB instance property in JSON format and parses it
 *
 * v 1.0.3
 * read ALL collections from a database in BSON format AND
 * store them into a Storj bucket
 *
 * v 1.0.2
 * read ALL collections from a database in BSON format
 *
 * v 1.0.1
 * provides option to connect to a MongoDB instance,
 * whose properties are read from given file
 *
 * v 1.0.0
 * provides options to read following files:
 * 	- MongoDB property
 *	- Storj Configuration information in JSON format
 */

package main

import (
	"encoding/json"
    "fmt"
    "github.com/urfave/cli"
    "io/ioutil"
    "log"
    "mongo"
    "os"
    "storj"
    "time"
    "unsafe"
)	

var gb_DEBUG bool

// create command-line tool
var app = cli.NewApp()

// setAppInfo sets information about the command-line application
func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to decentralized Storj cloud"
	app.Author = "Satyam Shivam" 
	app.Version = "1.0.5"
}

// setCommands sets various command-line options for the application
func setCommands() {

	app.Commands = []cli.Command {
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read and parse JSON information about MongoDB instance properties and then fetch ALL its collections\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_connector.json\n\n     example = ./storj_mongodb d ./config/db_property.json\n\n\n",
			Action: func(cliContext *cli.Context) { 
				var fullFilename string = "./config/db_property.json"
				//
				if (len(cliContext.Args()) > 0) {
					for i := 0; i < len(cliContext.Args()); i++ {
						// if any of the arguments involve "debug"
						if (cliContext.Args()[i] == "debug") {
							gb_DEBUG = true
							mongo.DEBUG = gb_DEBUG
							storj.DEBUG = gb_DEBUG
						} else {
							fullFilename = cliContext.Args()[i]
						}
					}	
				}

				data, dbname, err := mongo.ConnectToDB_FetchData(fullFilename)
				//	
				if err != nil {
					log.Fatalf("mongo.connectToDB_FetchData: %s", err)
				} else {
					fmt.Println("\nReading ALL collections from the MongoDB database...Complete!")
				}

				if gb_DEBUG {
					fmt.Println("Size of fetched data from database :", dbname, unsafe.Sizeof(data));
				}
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network and upload a sample BSON data\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) { 
				// default Storj configuration file name
				var fullFileName string = "./config/storj_config.json"
				//
				if (len(cliContext.Args()) > 0) {
					for i := 0; i < len(cliContext.Args()); i++{
						// if any of the arguments involve "debug"
						if (cliContext.Args()[i] == "debug") {
							gb_DEBUG = true
							mongo.DEBUG = gb_DEBUG
							storj.DEBUG = gb_DEBUG
						} else {
							fullFileName = cliContext.Args()[i]
						}
					}	
				}
				// sample database name
				dbName := "testdb"
				//
				// sample data to be uploaded
				jsonData := "{'testKey': 'testValue'}"
				//
				// Converting JSON data to bson data 
				bsonData, _ := json.Marshal(jsonData)

				if (gb_DEBUG) {
					t := time.Now()
    				time := t.Format("2006-01-02_15:04:05")
					var filename string = "uploaddata_" + time + ".bson"
					//
					err := ioutil.WriteFile(filename, bsonData, 0644)
					if err != nil {
						fmt.Println("Error while writting to file ")
					}
				}

				err := storj.ConnectStorj_UploadData(fullFileName, []byte(bsonData), dbName)
				if err != nil {
					fmt.Println("Error while uploading data to the Storj bucket")
				}
			},
		},
		{
			Name : "connect",
			Aliases : []string{"c"},
			Usage : "use this command to connect and transfer ALL collections from a desired MongoDB instance to given Storj Bucket in BSON format\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties in JSON format\n\n       if this filename is not given, then data is read from ./config/db_property.json\n\n       2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb c ./config/db_property.json ./config/storj_config.json\n\n\n",
			Action : func(cliContext *cli.Context) {
				// default Storj configuration file name
				var fullFileNameStorj string = "./config/storj_config.json"
				//
				var fullFileNameMongodb string = "./config/db_property.json"
				// reading filename from the command line 
				var firstfilename bool = false
				if (len(cliContext.Args()) > 0) {
					for i := 0; i < len(cliContext.Args()); i++ {
						// if any of the arguments involve "debug"
						if (cliContext.Args()[i] == "debug") {							
							gb_DEBUG = true
							mongo.DEBUG = gb_DEBUG
							storj.DEBUG = gb_DEBUG
						} else {
							if (!firstfilename) {
								fullFileNameMongodb = cliContext.Args()[i]
								firstfilename = true
							} else {
								fullFileNameStorj = cliContext.Args()[i]
							}
						}
					}	
				}
				//
				// fetching data from the mongodb
				data, dbname, err := mongo.ConnectToDB_FetchData(fullFileNameMongodb)	
				//
				if err != nil {
					log.Fatalf("mongo.connectToDB_FetchData: %s", err)
				}
				//
				if gb_DEBUG {
					fmt.Println("Size of fetched data from database: ", dbname, unsafe.Sizeof(data))
				}
				// connecting to storj network for uploading data
				err = storj.ConnectStorj_UploadData(fullFileNameStorj, []byte(data), dbname)
				if err != nil {
					fmt.Println("Error while uploading data to bucket ", err)
				}
			},
		},
	}
}


func main() {

	setAppInfo()
	setCommands()

	gb_DEBUG = false
	mongo.DEBUG = gb_DEBUG
	storj.DEBUG = gb_DEBUG

	err := app.Run(os.Args)
	//
	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}