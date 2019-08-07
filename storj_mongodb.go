// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"encoding/json"
	"os"
	"fmt"
	"log"
	"time"
	"unsafe"
	"io/ioutil"
	
	"storj"
	"mongo"
	"github.com/urfave/cli"
)

var gbDEBUG bool

// Create command-line tool to read from CLI.
var app = cli.NewApp()

// SetAppInfo sets information about the command-line application.
func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to decentralized Storj cloud"
	app.Author = "Satyam Shivam"
	app.Version = "1.0.5"
}

// setCommands sets various command-line options for the app.
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
						// Incase, debug is provided as argument.
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
				// Default Storj configuration file name.
				var fullFileName = "./config/storj_config.json"
				//
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase, debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							gbDEBUG = true
							mongo.DEBUG = gbDEBUG
							storj.DEBUG = gbDEBUG
						} else {
							fullFileName = cliContext.Args()[i]
						}
					}
				}
				// Sample database name.
				dbName := "testdb"
				//
				// Sample data to be uploaded.
				jsonData := "{'testKey': 'testValue'}"
				//
				// Converting JSON data to bson data.
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
				// Default Storj configuration file name.
				var fullFileNameStorj = "./config/storj_config.json"
				//
				var fullFileNameMongoDB = "./config/db_property.json"
				// Reading fileName from the command line.
				var foundFirstFileName = false
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase, debug is provided as argument.
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
				// Fetching data from the mongodb.
				data, dbname, err := mongo.ConnectToDBFetchData(fullFileNameMongoDB)
				//
				if err != nil {
					log.Fatalf("mongo.ConnectToDBFetchData: %s", err)
				}
				//
				if gbDEBUG {
					fmt.Println("Size of fetched data from database: ", dbname, unsafe.Sizeof(data))
				}
				// Connecting to storj network for uploading data.
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
