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
	"io/ioutil"
	"log"
	"mongo"
	"os"
	"storj"
	"time"
	"unsafe"

	"github.com/urfave/cli"
)

var gbDEBUG bool

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

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read and parse JSON information about MongoDB instance properties and then fetch ALL its collections\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this fileName is not given, then data is read from ./config/db_connector.json\n\n     example = ./storj_mongodb d ./config/db_property.json\n\n\n",
			Action: func(cliContext *cli.Context) {
				var fullFileName = "./config/db_property.json"
				//
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// if any of the arguments involve "debug"
						if cliContext.Args()[i] == "debug" {
							gbDEBUG = true
							mongo.DEBUG = gbDEBUG
							storj.DEBUG = gbDEBUG
						} else {
							fullFileName = cliContext.Args()[i]
						}
					}
				}

				data, dbname, err := mongo.ConnectToDBFetchData(fullFileName)
				//
				if err != nil {
					log.Fatalf("mongo.ConnectToDBFetchData: %s", err)
				} else {
					fmt.Println("\nReading ALL collections from the MongoDB database...Complete!")
				}

				if gbDEBUG {
					fmt.Println("Size of fetched data from database :", dbname, unsafe.Sizeof(data))
				}
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network and upload a sample BSON data\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this fileName is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) {
				// default Storj configuration file name
				var fullFileName = "./config/storj_config.json"
				//
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// if any of the arguments involve "debug"
						if cliContext.Args()[i] == "debug" {
							gbDEBUG = true
							mongo.DEBUG = gbDEBUG
							storj.DEBUG = gbDEBUG
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

				if gbDEBUG {
					t := time.Now()
					time := t.Format("2006-01-02_15:04:05")
					var fileName = "uploaddata_" + time + ".bson"
					//
					err := ioutil.WriteFile(fileName, bsonData, 0644)
					if err != nil {
						fmt.Println("Error while writting to file ")
					}
				}

				err := storj.ConnectStorjUploadData(fullFileName, []byte(bsonData), dbName)
				if err != nil {
					fmt.Println("Error while uploading data to the Storj bucket")
				}
			},
		},
		{
			Name:    "connect",
			Aliases: []string{"c"},
			Usage:   "use this command to connect and transfer ALL collections from a desired MongoDB instance to given Storj Bucket in BSON format\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties in JSON format\n\n       if this fileName is not given, then data is read from ./config/db_property.json\n\n       2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n\n       if this fileName is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb c ./config/db_property.json ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) {
				// default Storj configuration file name
				var fullFileNameStorj = "./config/storj_config.json"
				//
				var fullFileNameMongoDB = "./config/db_property.json"
				// reading fileName from the command line
				var foundFirstFileName = false
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// if any of the arguments involve "debug"
						if cliContext.Args()[i] == "debug" {
							gbDEBUG = true
							mongo.DEBUG = gbDEBUG
							storj.DEBUG = gbDEBUG
						} else {
							if !foundFirstFileName {
								fullFileNameMongoDB = cliContext.Args()[i]
								foundFirstFileName = true
							} else {
								fullFileNameStorj = cliContext.Args()[i]
							}
						}
					}
				}
				//
				// fetching data from the mongodb
				data, dbname, err := mongo.ConnectToDBFetchData(fullFileNameMongoDB)
				//
				if err != nil {
					log.Fatalf("mongo.ConnectToDBFetchData: %s", err)
				}
				//
				if gbDEBUG {
					fmt.Println("Size of fetched data from database: ", dbname, unsafe.Sizeof(data))
				}
				// connecting to storj network for uploading data
				err = storj.ConnectStorjUploadData(fullFileNameStorj, []byte(data), dbname)
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

	gbDEBUG = false
	mongo.DEBUG = gbDEBUG
	storj.DEBUG = gbDEBUG

	err := app.Run(os.Args)
	//
	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
