/*
 * package to read MongoDB collections into a Storj bucket
 *
 * v 1.0.0
 * provides options to read following files:
 * 	- MongoDB property
 *	- Storj Configuration information in  JSON format
 */

package main

import (
    "fmt"
    "os"
    "encoding/json"
    "github.com/urfave/cli"
	"log"
	"bufio"
)

// readPropertiesDB reads the database property file
// and returns all the credentials information
func readPropertiesDB(fullFileName string) ([]string, error) {
	fileHandle, err := os.Open(fullFileName)
	defer fileHandle.Close()
	if err != nil {
		return nil, err
	}
	
	var properties []string
	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		properties = append(properties, scanner.Text())
	}
	return properties, scanner.Err()
}

// ConfigStorj depicts keys to search for
// within the stroj_config.json file
type ConfigStorj struct { 
	ApiKey     string `json:"apikey"`
	Satellite  string `json:"satellite"`
	Bucket     string `json:"bucket"`
	UploadPath string `json:"uploadPath"`
}

// LoadStorjConfiguration reads and parses the JSON file
// that contain Storj configuration information
func LoadStorjConfiguration(fullFileName string) (ConfigStorj, error) {
	var configStorj ConfigStorj
	
	fileHandle, err := os.Open(fullFileName)
	defer fileHandle.Close()
	if err != nil {
		return configStorj, err
		// fmt.Println(err.Error())
	}
	
	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)
	
	return configStorj, nil
}

// create command-line tool
var app = cli.NewApp()

func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to decentralized Storj cloud"
	app.Author = "Satyam Shivam" 
	app.Version = "1.0.0"
}

func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read properties of desired mongo database\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB credentials\n\n       if this filename is not given, then data is read from ./config/db_connector\n\n     example =   ./storj_mongodb d ./config/db_property\n\n\n",
			Action: func(c *cli.Context) { 

				// default full file name
				var fullFileName string="./config/db_property"

				if len(c.Args())>0 {
					fullFileName = c.Args()[0]
				}

				mongoDbProperty, err := readPropertiesDB(fullFileName)
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				
				// print to debug
				fmt.Println("Read MongoDB credentials from the ", fullFileName, " file")
				fmt.Println("Host Name	: ", mongoDbProperty[0])
				fmt.Println("Port Number\t: ", mongoDbProperty[1])
				fmt.Println("User Name	: ", mongoDbProperty[2])
				fmt.Println("Password	: ", mongoDbProperty[3])
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example =   ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(c *cli.Context) { 

				// default full file name
				var fullFileName string="./config/storj_config.json"

				if len(c.Args())>0 {
					fullFileName = c.Args()[0]
				}

				configStorj, err := LoadStorjConfiguration(fullFileName)
				if err != nil {
					log.Fatalf("LoadStorjConfiguration: %s", err)
				}

				// print to debug
				fmt.Println("Read Storj configuration from the ", fullFileName, " file")
				fmt.Println("API Key\t\t: ", configStorj.ApiKey)
				fmt.Println("Satellite	: ", configStorj.Satellite)
				fmt.Println("Bucket		: ", configStorj.Bucket)
				fmt.Println("Upload Path\t: ", configStorj.UploadPath)
			},
		},
	}
}


func main() {

	setAppInfo()
	setCommands()

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}