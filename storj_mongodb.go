/*
 * package to read MongoDB collections into a Storj bucket
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
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "github.com/urfave/cli"
	"log"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConfigMongoDB define variables and types
type ConfigMongoDB struct {

	fullFileName	string
	//
	hostName 		string
	portNumber		string
	userName 		string 
	password		string

}

var gO_configMongoDB ConfigMongoDB

// readPropertiesDB asks for the name of a file
// wherefrom it reads the database property contents
// and returns all the credentials information as an object
func readPropertiesDB(cliContext *cli.Context) (ConfigMongoDB, error) {
	var configMongoDB ConfigMongoDB
	
	// default full file name
	configMongoDB.fullFileName = "./config/db_property"

	if len(cliContext.Args()) > 0 {
		configMongoDB.fullFileName = cliContext.Args()[0]
	}

	// open the desired file
	fileHandle, err := os.Open(configMongoDB.fullFileName)
	if err != nil {
		return configMongoDB, err
	}
	defer fileHandle.Close()
	
	// read contents of the file
	var dbProperty []string
	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		dbProperty = append(dbProperty, scanner.Text())
	}
	//
	err = scanner.Err()
	if err != nil {
		return configMongoDB, err
	}

	// store contents into the structure elements 
	configMongoDB.hostName 	 = dbProperty[0]
	configMongoDB.portNumber = dbProperty[1]
	configMongoDB.userName   = dbProperty[2]
	configMongoDB.password 	 = dbProperty[3]
	
	return configMongoDB, nil
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
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()
	
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
	app.Version = "1.0.1"
}

func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read properties of desired mongo database\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB credentials\n\n       if this filename is not given, then data is read from ./config/db_connector\n\n     example =   ./storj_mongodb d ./config/db_property\n\n\n",
			Action: func(cliContext *cli.Context) { 

				gO_configMongoDB, err := readPropertiesDB(cliContext)
				//
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				
				// print to debug
				fmt.Println("Read MongoDB credentials from the ", gO_configMongoDB.fullFileName, " file")
				fmt.Println("Host Name	: ", gO_configMongoDB.hostName)
				fmt.Println("Port Number\t: ", gO_configMongoDB.portNumber)
				fmt.Println("User Name	: ", gO_configMongoDB.userName)
				fmt.Println("Password	: ", gO_configMongoDB.password)
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example =   ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) { 

				// default full file name
				var fullFileName string="./config/storj_config.json"

				if len(cliContext.Args()) > 0 {
					fullFileName = cliContext.Args()[0]
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
		{
			Name : "connect",
			Aliases : []string{"c"},
			Usage : "use this command to connect with a mongoDB instance\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB credentials\n\n       if this filename is not given, then data is read from ./config/db_property\n\n     example =   ./storj_mongodb c ./config/db_property\n\n\n",
			Action : func(cliContext *cli.Context){
				// read from a file
				var err error
				gO_configMongoDB, err = readPropertiesDB(cliContext)
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				
				// print to debug
				fmt.Println("Read MongoDB credentials from the ", gO_configMongoDB.fullFileName, " file")
				fmt.Println("Host Name	: ", gO_configMongoDB.hostName)
				fmt.Println("Port Number\t: ", gO_configMongoDB.portNumber)
				fmt.Println("User Name	: ", gO_configMongoDB.userName)
				fmt.Println("Password	: ", gO_configMongoDB.password)


				// connection to mongoDB
				mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:%s/admin?authSource=admin", gO_configMongoDB.userName, gO_configMongoDB.password, gO_configMongoDB.hostName, gO_configMongoDB.portNumber)
				clientOptions := options.Client().ApplyURI(mongoURL)
				client, err := mongo.Connect(context.TODO(), clientOptions)

				if err != nil {
					log.Fatal(err)
				}

				// Check the connection
				err = client.Ping(context.TODO(), nil)
				if err != nil {
					log.Fatal(err)
				}

				// print if connection successfully
				fmt.Println("Connected to MongoDB!")
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