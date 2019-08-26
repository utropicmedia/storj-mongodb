// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	"unsafe"

	"utropicmedia/mongodb_storj_interface/mongo"
 	"utropicmedia/mongodb_storj_interface/storj"

	"github.com/urfave/cli"
)

const dbConfigFile = "./config/db_property.json"
const storjConfigFile = "./config/storj_config.json"

var gbDEBUG = false

// Create command-line tool to read from CLI.
var app = cli.NewApp()

// SetAppInfo sets information about the command-line application.
func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to the decentralized Storj network"
	app.Author = "Satyam Shivam - Utropicmedia"
	app.Version = "1.0.9"

}

// helper function to flag debug
func setDebug(debugVal bool) {
	gbDEBUG = debugVal
	mongo.DEBUG = debugVal
	storj.DEBUG = debugVal
}

// setCommands sets various command-line options for the app.
func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "parse",
			Aliases: []string{"p"},
			Usage:   "Command to read and parse JSON information about MongoDB instance properties and then fetch ALL its collections. ",
			//\narguments-\n\t  fileName [optional] = provide full file name (with complete path), storing mongoDB properties if this fileName is not given, then data is read from ./config/db_connector.json\n\t  example = ./storj_mongodb d ./config/db_property.json\n",
			Action: func(cliContext *cli.Context) {
				var fullFileName = dbConfigFile

				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args()[i]
						}
					}
				}

				// Establish connection with MongoDB and get io.Reader implementor.
				dbReader, err := mongo.ConnectToDB(fullFileName)
				//
				if err != nil {
					log.Fatalf("Failed to establish connection with MongoDB: %s\n", err)
				}

				// Connect to the Database and process data
				data, err := mongo.FetchData(dbReader)

				if err != nil {
					log.Fatalf("mongo.FetchData: %s", err)
				} else {
					fmt.Println("Reading ALL collections from the MongoDB database...Complete!")
				}

				if gbDEBUG {
					fmt.Println("Size of fetched data from database :", dbReader.DatabaseName, unsafe.Sizeof(data))
				}
			},
		},
		{
			Name:    "test",
			Aliases: []string{"t"},
			Usage:   "Command to read and parse JSON information about Storj network and upload sample JSON data",
			//\n arguments- 1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information if this fileName is not given, then data is read from ./config/storj_config.json example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) {

				// Default Storj configuration file name.
				var fullFileName = storjConfigFile

				// process arguments
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {

						// Incase, debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
						} else {
							fullFileName = cliContext.Args()[i]
						}
					}
				}

				// Sample database name and data to be uploaded
				dbName := "testdb"
				jsonData := "{'testKey': 'testValue'}"

				// Converting JSON data to bson data.  TODO: convert to BSON using call to mongo library
				bsonData, _ := json.Marshal(jsonData)

				if gbDEBUG {
					t := time.Now()
					time := t.Format("2006-01-02_15:04:05")
					var fileName = "uploaddata_" + time + ".bson"

					err := ioutil.WriteFile(fileName, bsonData, 0644)
					if err != nil {
						fmt.Println("Error while writting to file ")
					}
				}

				// Create a buffer as an io.Reader implementor.
				buf := bytes.NewBuffer(bsonData)
				//
				err := storj.ConnectStorjReadUploadData(fullFileName, buf, dbName)
				//
				if err != nil {
					fmt.Println("Error while uploading data to the Storj bucket")
				}
			},
		},
		{
			Name:    "store",
			Aliases: []string{"s"},
			Usage:   "Command to connect and transfer ALL collections from a desired MongoDB instance to given Storj Bucket in BSON format",
			//\n    arguments-\n      1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties in JSON format\n   if this fileName is not given, then data is read from ./config/db_property.json\n      2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n     if this fileName is not given, then data is read from ./config/storj_config.json\n   example = ./storj_mongodb c ./config/db_property.json ./config/storj_config.json\n",
			Action: func(cliContext *cli.Context) {

				// Default configuration file names.
				var fullFileNameStorj = storjConfigFile
				var fullFileNameMongoDB = dbConfigFile

				// process arguments - Reading fileName from the command line.
				var foundFirstFileName = false
				if len(cliContext.Args()) > 0 {
					for i := 0; i < len(cliContext.Args()); i++ {
						// Incase debug is provided as argument.
						if cliContext.Args()[i] == "debug" {
							setDebug(true)
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

				// Establish connection with MongoDB and get io.Reader implementor.
				dbReader, err := mongo.ConnectToDB(fullFileNameMongoDB)
				//
				if err != nil {
					log.Fatalf("Failed to establish connection with MongoDB: %s\n", err)
				}

				// Fetch all collections' documents from MongoDB instance
				// and simultaneously store them into desired Storj bucket.
				err = storj.ConnectStorjReadUploadData(fullFileNameStorj, dbReader, dbReader.DatabaseName)
				//
				if err != nil {
					log.Fatalf("Error while fetching MongoDB documents and uploading them to bucket: %s\n", err)
				}
			},
		},
	}
}

func main() {

	setAppInfo()
	setCommands()

	setDebug(false)

	err := app.Run(os.Args)

	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
