// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package mongo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	goMongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false


// ConfigMongoDB defines the variables and types.
type ConfigMongoDB struct {
	Hostname   string `json:"hostname"`
	Portnumber string `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Database   string `json:"database"`
}

// MongoReader implements an io.Reader interface
type MongoReader struct {
	DatabaseName		string
	database 			*goMongo.Database
	collectionNames 	[]string
	lastDocumentIndex	int
}

// Read reads and copies as many collections' documents raw BSON data into the 
// buffer, as per the capacity of the buffer 
func (mongoReader *MongoReader) Read(buf []byte) (int, error) { // buf represents the byte array, where data is to be copied
	// It returns number of bytes (int) that are copied
	// and any error, if occurred.
	// At the end of complete reading, io.EOF is sent as part of the error.
	var err error
	ctx := context.TODO()

	bufferCapacity := cap(buf)

	filterBSON := bson.M{}
	//
	if mongoReader.collectionNames == nil {
		fmt.Println("Reading ALL collections from the MongoDB database...")
		
		// Retrieve ALL collections in the database.
		mongoReader.collectionNames, err = mongoReader.database.ListCollectionNames(ctx, filterBSON)
		//
		if err != nil {
			log.Printf("Failed to retrieve collection names: %s\n", err)
			return 0, err
		}
	} else {
		fmt.Printf("Reading from MongoDB collection %s...\n", mongoReader.collectionNames[0])
	}
	
	var numOfBytesRead = 0

	// Go through ALL collections.
	for ij, _ := range mongoReader.collectionNames {
		if DEBUG {
			fmt.Println("Collection: ", mongoReader.collectionNames[ij])
			fmt.Println("-----------------")
		}

		collection := mongoReader.database.Collection(mongoReader.collectionNames[ij])
		//
		cursor, err := collection.Find(ctx, filterBSON)
		//
		if err != nil {
			log.Printf("Failed to retrieve data about %s collection: %s\n", mongoReader.collectionNames[ij], err)
			return numOfBytesRead, err
		}
		//
		defer cursor.Close(ctx)
		
		var documentCount = 0
		
		// Retrieve each document of the selected collection.
		for cursor.Next(ctx) {
			// Start reading from the last document that was under process.
			if documentCount >= mongoReader.lastDocumentIndex {
				// Convert JSON to BSON.
				rawDocumentBSON, _ := bson.Marshal(cursor.Current)
				//
				documentSize := len(rawDocumentBSON)
				//
				// Ensure required space is available in buf.
				if (numOfBytesRead + documentSize) < bufferCapacity {
					copy(buf[numOfBytesRead:], rawDocumentBSON)
					//
					numOfBytesRead += documentSize	
				} else {
					if DEBUG {
						log.Printf("Buffer full with %d bytes of '%s' collection's data from documents #%d to #%d!\nCaller must recall the Reader.\n", numOfBytesRead, mongoReader.collectionNames[ij], mongoReader.lastDocumentIndex, documentCount)
					}
					// Insufficient space in the buffer.
					err = io.ErrShortBuffer
	
					// Next time, start with the unprocessed collections.
					mongoReader.collectionNames = mongoReader.collectionNames[ij:]
					//
					// Also, operate from the current document instance.
					mongoReader.lastDocumentIndex = documentCount
	
					// More data still needs to be sent,
					// hence, caller must recall the Reader.
					return numOfBytesRead, err
				}
			}
			//
			documentCount++
		}

		if DEBUG {
			fmt.Printf("Retrieved %d bytes of '%s' collection's data from documents #%d to #%d", numOfBytesRead, mongoReader.collectionNames[ij], mongoReader.lastDocumentIndex, documentCount)
		}
		
		if err := cursor.Err(); err != nil {
			if DEBUG {
				fmt.Printf(" before error: %s.\n", err)
			}
			// Unexpected error occurred while processing cursors.
			// Next time, start with the unprocessed collections.
			mongoReader.collectionNames = mongoReader.collectionNames[ij:]
			//
			// Also, operate from the current document instance.
			mongoReader.lastDocumentIndex = documentCount

			// The caller needs to recall the Reader to fetch left-over data.
			return numOfBytesRead, err
		}

		if DEBUG {
			fmt.Printf(".\n")
		}
		
		log.Println("ALL documents of the collection are read!")

		// All documents of the selected collection have been read.
		if mongoReader.lastDocumentIndex > 0 {
			// Reset the document index to be read from.
			mongoReader.lastDocumentIndex = 0
		}
	}

	// All collections have been read and processed.
	mongoReader.collectionNames = nil

	return numOfBytesRead, io.EOF
}

// LoadMongoProperty reads and parses the JSON file.
// that contain a MongoDB instance's property.
// and returns all the properties as an object.
func LoadMongoProperty(fullFileName string) (ConfigMongoDB, error) { // fullFileName for fetching database credentials from  given JSON filename.
	var configMongoDB ConfigMongoDB
	// Open and read the file
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configMongoDB, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configMongoDB)

	// Display read information.
	fmt.Println("Read MongoDB configuration from the ", fullFileName, " file")
	fmt.Println("Hostname\t", configMongoDB.Hostname)
	fmt.Println("Portnumber\t", configMongoDB.Portnumber)
	fmt.Println("Username \t", configMongoDB.Username)
	fmt.Println("Password \t", configMongoDB.Password)
	fmt.Println("Database \t", configMongoDB.Database)

	return configMongoDB, nil
}

// ConnectToDB will connect to a MongoDB instance,
// based on the read property from an external file.
// It returns a reference to an io.Reader with MongoDB instance information
func ConnectToDB(fullFileName string) (*MongoReader, error) { // fullFileName for fetching database credentials from given JSON filename.
	// Read MongoDB instance's properties from an external file.
	configMongoDB, err := LoadMongoProperty(fullFileName)
	//
	if err != nil {
		log.Printf("LoadMongoProperty: %s\n", err)
		return nil, err
	}

	fmt.Println("Connecting to MongoDB...")

	mongoURL := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource="+configMongoDB.Database, configMongoDB.Username, configMongoDB.Password, configMongoDB.Hostname, configMongoDB.Portnumber, configMongoDB.Database)
	//
	clientOptions := options.Client().ApplyURI(mongoURL)
	//
	client, err := goMongo.Connect(context.TODO(), clientOptions)
	//
	if err != nil {
		log.Printf("mongo.Connect: %s\n", err)
		return nil, err
	}

	// Check the connection with MongoDB.
	err = client.Ping(context.TODO(), nil)
	//
	if err != nil {
		fmt.Println("FAILed to connect to the MongoDB instance!")
		log.Printf("client.Ping: %s\n", err)
		return nil, err
	}

	// Inform about successful connection.
	fmt.Println("Successfully connected to MongoDB!")

	return &MongoReader{DatabaseName: configMongoDB.Database, database: client.Database(configMongoDB.Database)}, nil
}

// FetchData reads ALL collections' BSON data, and
// returns them in appended format.
func FetchData(databaseReader io.Reader) ([]byte, error) { // databaseReader is an io.Reader implementation that 'reads' desired data.
	// Create a buffer of feasible size
	rawDocumentBSON := make([]byte, 0, 32768)
	
	// Retrieve ALL collections in the database.
	var allCollectionsDataBSON = []byte{}
	
	var numOfBytesRead int
	var err error

	// Read data using the given io.Reader.
	for err = io.ErrShortBuffer; (err == io.ErrShortBuffer); {
		numOfBytesRead, err = databaseReader.Read(rawDocumentBSON)
		//
		if numOfBytesRead > 0 {
			// Append the BSON data to earlier one.
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], rawDocumentBSON...)
			//
			if DEBUG {
				fmt.Printf("Read %d bytes of data - Error: %s == %s => %t\n", numOfBytesRead, err, io.ErrShortBuffer, err == io.ErrShortBuffer)
			}	
		}
	}
	//
	if DEBUG {
		// complete BSON data from ALL collections.
		t := time.Now()
		time := t.Format("2006-01-02_15:04:05")
		var filename = "uploaddata_" + time + ".bson"
		err = ioutil.WriteFile(filename, allCollectionsDataBSON, 0644)
	}

	return allCollectionsDataBSON, err
}
