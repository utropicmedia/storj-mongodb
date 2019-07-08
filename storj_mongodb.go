/*
 * package to read MongoDB collections into a Storj bucket
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
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "log"
	"os"
	"time"

	"github.com/urfave/cli"
    "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	
)

var gb_DEBUG bool

// ConfigMongoDB define variables and types
type ConfigMongoDB struct {

	fullFileName	string
	//
	hostName 		string
	portNumber		string
	userName 		string 
	password		string
	//
	database 		string

}

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
	configMongoDB.database 	 = dbProperty[4]
	
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
	app.Version = "1.0.2"
}

func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read properties of desired mongo database\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_connector\n\n     example = ./storj_mongodb d ./config/db_property\n\n\n",
			Action: func(cliContext *cli.Context) { 

				configMongoDB, err := readPropertiesDB(cliContext)
				//
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				
				// print to debug
				fmt.Println("Read MongoDB credentials from the ", configMongoDB.fullFileName, " file")
				fmt.Println("Host Name	: ", configMongoDB.hostName)
				fmt.Println("Port Number\t: ", configMongoDB.portNumber)
				fmt.Println("User Name	: ", configMongoDB.userName)
				fmt.Println("Password	: ", configMongoDB.password)
				fmt.Println("Database	: ", configMongoDB.database)
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
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
			Usage : "use this command to connect with a mongoDB instance\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_property\n\n     example = ./storj_mongodb c ./config/db_property\n\n\n",
			Action : func(cliContext *cli.Context){
				// read MongoDB's properties, credentials, and database name from a file
				configMongoDB, err := readPropertiesDB(cliContext)
				//
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}				
				
				// display read information about MongoDB properties, etc.
				fmt.Println("Read MongoDB credentials from the ", configMongoDB.fullFileName, " file")
				fmt.Println("Host Name	: ", configMongoDB.hostName)
				fmt.Println("Port Number\t: ", configMongoDB.portNumber)
				fmt.Println("User Name	: ", configMongoDB.userName)
				fmt.Println("Password	: ", configMongoDB.password)
				fmt.Println("Database	: ", configMongoDB.database)
				
				
				// Connect to MongoDB
				fmt.Println("\nConnecting to MongoDB...")

				mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", configMongoDB.userName, configMongoDB.password, configMongoDB.hostName, configMongoDB.portNumber, configMongoDB.database)
				// 
				clientOptions := options.Client().ApplyURI(mongoURL)
				//
				client, err := mongo.Connect(context.TODO(), clientOptions)

				if err != nil {
					log.Fatal(err)
				}

				// Check the connection
				err = client.Ping(context.TODO(), nil)
				if err != nil {
					log.Fatal(err)
				}

				// Inform about successful connection
				fmt.Println("Connected to MongoDB!")

				fmt.Println("\nReading ALL Collections from the MongoDB database...")

				// Read database name from the db_property file
				db := client.Database(configMongoDB.database)
				
				// Create a new context with a 10 second timeout
				ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
				defer cancel()

				// Retrieve ALL collections in the database
				filter := bson.M{}

				collectionNames, err:= db.ListCollectionNames(ctx, filter)
				if err != nil {
				    // Handle error
				    log.Printf("Failed to get collections: %v", err)
				    return
				}

				// Go through ALL collections
				for _, collectionName := range collectionNames {
					if gb_DEBUG { 
						fmt.Println("\nCollection: ", collectionName) 
						fmt.Println("-----------------") 
					}
					
				    collection := db.Collection(collectionName)
				    
				    cursor, err := collection.Find(ctx, filter)
				    //
				    if err != nil {
						log.Fatal(err)
					}

					// Retrieve ALL documents from the selected collection
					for cursor.Next(ctx) {
						// JSON document
						rawDocumentJSON := cursor.Current
						// Convert JSON to BSON
						rawDocumentBSON, _ := bson.Marshal(rawDocumentJSON)
						fmt.Println("\nBSON", rawDocumentBSON)

						if gb_DEBUG {
							// DE-BUG
							var bson2JSON interface{}
							err = bson.Unmarshal(rawDocumentBSON, &bson2JSON)
							fmt.Println("JSON equivalent: ", bson2JSON)
							//
							if err != nil {
								fmt.Println(err)
							}
						}
				    }
				}	
			},
		},
	}
}


func main() {

	setAppInfo()
	setCommands()

	gb_DEBUG = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}