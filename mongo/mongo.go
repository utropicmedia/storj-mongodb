/*
 * mongo package connects to a Mongo database instance,
 * based on the properties, read from a JSON file.
 * It then reads BSON data from all the collections.
 *
 * v 1.0.0
 * MongoDB functions collected into a separate package 
 */

package mongo

import(
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
	
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)


var DEBUG bool = false
var gb_DEBUG_DEV bool = false

// ConfigMongoDB define variables and types
type ConfigMongoDB struct {
	Hostname 	string	`json:"hostname"`
	Portnumber	string	`json:"port"`
	Username 	string	`json:"username"`
	Password	string	`json:"password"`
	Database 	string	`json:"database"`
}


// LoadMongoProperty reads and parses the JSON file
// that contain a MongoDB instance's property
// and returns all the properties as an object
func LoadMongoProperty(fullFileName string) (ConfigMongoDB, error) {
	var configMongoDB ConfigMongoDB
	// Open and read the file
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configMongoDB, err
	}
	defer fileHandle.Close()
	
	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configMongoDB)

	// Display read information
	fmt.Println("Read MongoDB configuration from the ", fullFileName, " file")
	fmt.Println("Hostname\t", configMongoDB.Hostname)
	fmt.Println("Portnumber\t", configMongoDB.Portnumber)
	fmt.Println("Username \t", configMongoDB.Username)
	fmt.Println("Password \t", configMongoDB.Password)
	fmt.Println("Database \t", configMongoDB.Database)

	return configMongoDB, nil
}


// ConnectToDB_FetchData will connect to a MongoDB instance,
// based on the read property from an external file
// it then reads ALL collections' BSON data, and 
// returns them in appended format 
func ConnectToDB_FetchData(fullFileName string) ([]byte, string, error) {
	// Read MongoDB instance's properties from an external file
	configMongoDB, err := LoadMongoProperty(fullFileName)
	if err != nil {
		log.Fatalf("loadMongoProperty: %s", err)
	}


	fmt.Println("\nConnecting to MongoDB...")

	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=admin", configMongoDB.Username, configMongoDB.Password, configMongoDB.Hostname, configMongoDB.Portnumber, configMongoDB.Database)
	// 
	clientOptions := options.Client().ApplyURI(mongoURL)
	//
	client, err := mongo.Connect(context.TODO(), clientOptions)
	//
	if err != nil {
		log.Fatalf("mongo.Connect: %s", err)
	}

	// Check the connection with MongoDB
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		fmt.Println("FAILed to connect to the MongoDB instance!")
		log.Fatalf("client.Ping: %s", err)
	}

	// Inform about successful connection
	fmt.Println("Successfully connected to MongoDB!")


	fmt.Println("\nReading ALL collections from the MongoDB database...")

	// Read database name from the db_property file
	db := client.Database(configMongoDB.Database)
	
	// Create a new context with a 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel()

	// Retrieve ALL collections in the database
	var allCollectionsDataBSON = []byte{}
	
	filterBSON := bson.M{}
	//
	collectionNames, err:= db.ListCollectionNames(ctx, filterBSON)
	if err != nil {
	    log.Printf("Failed to retrieve collection names: %s", err)
	    return allCollectionsDataBSON, "", err
	}

	// Go through ALL collections
	for _, collectionName := range collectionNames {
		if DEBUG { 
			fmt.Println("\nCollection: ", collectionName) 
			fmt.Println("-----------------") 
		}
		
	    collection := db.Collection(collectionName)
	    
	    cursor, err := collection.Find(ctx, filterBSON)
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
			// Append the BSON data to earlier one
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], rawDocumentBSON...)
	    }		
	}
	//
	if DEBUG {
		// complete BSON data from ALL collections
		t := time.Now()
    	time := t.Format("2006-01-02_15:04:05")
    	var filename string = "uploaddata_"+time+".bson"
		err = ioutil.WriteFile(filename, allCollectionsDataBSON, 0644)
		//
		// converting it into its equivalent JSON
	}

	return allCollectionsDataBSON, configMongoDB.Database, nil
}

