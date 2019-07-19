/*
 * package to read MongoDB collections into a Storj bucket
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
    "os"
    "log"
    "fmt"
    "time"    
	"context"
    "encoding/json"
    "github.com/urfave/cli"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    "storj.io/storj/lib/uplink"
	"bytes"
    "unsafe"
	"io/ioutil"
	"reflect"
)

var gb_DEBUG bool


// ConfigMongoDB define variables and types
type ConfigMongoDB struct {
	Hostname 	string	`json:"hostname"`
	Portnumber	string	`json:"port"`
	Username 	string	`json:"username"`
	Password	string	`json:"password"`
	Database 	string	`json:"database"`
}

// loadMongoProperty reads and parses the JSON file
// that contain a MongoDB instance's property
// and returns all the properties as an object
func loadMongoProperty(fullFileName string) (ConfigMongoDB, error) {
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

// connectMongo_FetchData will connect to a MongoDB instance,
// based on the read property from an external file
// it then reads ALL collections' BSON data, and 
// returns them in appended format 
func connectMongo_FetchData(cliContext *cli.Context) ([]byte, string,error) {
	// Read MongoDB instance's properties from an external file
	fullFileName := "./config/db_property.json"
	if len(cliContext.Args()) > 0 {
		fullFileName = cliContext.Args()[0]
	}
	//
	configMongoDB, err := loadMongoProperty(fullFileName)
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
	    return allCollectionsDataBSON, "",err
	}

	// Go through ALL collections
	for _, collectionName := range collectionNames {
		if gb_DEBUG { 
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
			err = ioutil.WriteFile("data.bson", rawDocumentBSON, 0644)
			fmt.Printf("%+v", rawDocumentBSON)
			// Append the BSON data to earlier one
			allCollectionsDataBSON = append(allCollectionsDataBSON[:], rawDocumentBSON...)
			
			if gb_DEBUG {
				var bson2JSON interface{}
				err = bson.Unmarshal(rawDocumentBSON, &bson2JSON)
				//
				
				fmt.Println("JSON data read: ",rawDocumentBSON)
				fmt.Println("JSON data read: ",bson2JSON)
				
			}
	    }		
	}
	//
	if gb_DEBUG {
		fmt.Println("JSON data read: ", allCollectionsDataBSON)
		var bson2JSON interface{}
		err = bson.Unmarshal(allCollectionsDataBSON, &bson2JSON)
		//
		if err != nil {
			log.Printf("From Mongo functions bson. Unmarshal: %s",  err)
			} else {
			//fmt.Println("JSON data read: ",bson2JSON)
		}
	}

	return allCollectionsDataBSON, configMongoDB.Database,nil
}

// ConfigStorj depicts keys to search for
// within the stroj_config.json file
type ConfigStorj struct { 
	ApiKey     string `json:"apikey"`
	Satellite  string `json:"satellite"`
	Bucket     string `json:"bucket"`
	UploadPath string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
}

// loadStorjConfiguration reads and parses the JSON file
// that contain Storj configuration information
func loadStorjConfiguration(fullFileName string) (ConfigStorj, error) {
	var configStorj ConfigStorj
	
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configStorj, err
	}
	defer fileHandle.Close()
	
	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configStorj)

	// Display read information
	fmt.Println("Read Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.ApiKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)

	return configStorj, nil
}

// connectStorj_UploadData reads Storj configuration from given file,
// connects to the desired Storj network, and
// uploads given object to the desired bucket
func connectStorj_UploadData(fullFileName string, dataToUpload []byte, databaseName string) error {
	
	// Read Storj bucket's configuration from an external file
	configStorj, err := loadStorjConfiguration(fullFileName)
	if err != nil {
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}
	
	if gb_DEBUG {
		fmt.Println("\nSize of BSON data being uploaded to Storj V3: ", unsafe.Sizeof(dataToUpload))
	}
	
	fmt.Println("\nCreating New Uplink...")
	
	var cfg uplink.Config
	ctx := context.Background()

    upl, err := uplink.NewUplink(ctx, &cfg)
    if err != nil {
        return fmt.Errorf("Could not create new Uplink object: %s", err)
    }
    defer upl.Close()

	
	fmt.Println("\nParsing the API key...")
	
	key, err := uplink.ParseAPIKey(configStorj.ApiKey)
	if err != nil {
		return fmt.Errorf("Could not parse API key: %s", err)
	}
	//
	if gb_DEBUG { 
		fmt.Println("API key \t   :", key)
		fmt.Println("Serialized API key :", key.Serialize())
	}


    fmt.Println("\nOpening Project...")
	
	proj, err := upl.OpenProject(ctx, configStorj.Satellite, key)
	//
	if err != nil {
        return fmt.Errorf("Could not open project: %s", err)
    }
    defer proj.Close()

	
	if gb_DEBUG {
		/*
		fmt.Println("Creating Bucket ", configStorj.Bucket)
		//
		_, err = proj.CreateBucket(ctx, configStorj.Bucket, nil)
		//
		fmt.Println("Create Bucket " ,err)
		if err != nil {
			return fmt.Errorf("Could not create bucket: %s", err)
		}
		
	
		fmt.Println("\nList of buckets:")
		//
		list := uplink.BucketListOptions {
			Direction: storj.Forward}
		//
		for {
			result, err := proj.ListBuckets(ctx, &list)
			if err != nil {
				return fmt.Errorf("proj.ListBuckets: %s", err)
			}
			//
			for _, bucket := range result.Items {
				fmt.Println("\n\tBucket: ", bucket.Name)
			}
			//
			if !result.More {
				break
			}
			//
			list = list.NextPage(result)
		}
		*/
	}

	// Creating an encryption key from encryption passphrase
    fmt.Println("\nGet encryption key from pass phrase...")
	
	encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
    if err != nil {
        return fmt.Errorf("Could not create encryption key: %s", err)
	}

	// Creating an encryption context
	access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)
	fmt.Println("\nEncryption access \t:",access)

	if gb_DEBUG {
		serializedAccess, err := access.Serialize()
		if err != nil {
			fmt.Println("Error Serialized key : ", err)	
		}
		//
		fmt.Println("Serialized access key\t:", serializedAccess)
	}

	
	fmt.Println("\nOpening Bucket...", configStorj.Bucket)
	
    // Open up the desired Bucket within the Project    
    bucket, err := proj.OpenBucket(ctx, configStorj.Bucket, access)
    //
    if err != nil {
        return fmt.Errorf("Could not open bucket %q: %s", configStorj.Bucket, err)
	}	
	defer bucket.Close()
		
    
	fmt.Println("\nGetting data into a buffer...")
	buf := bytes.NewBuffer(dataToUpload)

	if gb_DEBUG {
		fmt.Println("Upload: Type to the data:", reflect.TypeOf(dataToUpload))
		fmt.Println("Upload: Type to the buffer:", reflect.TypeOf(buf))
	}

	
	fmt.Println("\nCreating file name in the bucket, as per current time...")
	t := time.Now()
    time := t.Format("2006-01-02_15:04:05")
    var filename string = databaseName + "_" + time + ".bson"
    configStorj.UploadPath = configStorj.UploadPath + filename


	var opts uplink.UploadOptions

	if gb_DEBUG {
		//opts.Volatile.RedundancyScheme = cfgbucket.Volatile.RedundancyScheme
		//opts.Volatile.EncryptionParameters = cfgbucket.EncryptionParameters
	}
    
    if gb_DEBUG {
		fmt.Println("Size of data uploading: ", unsafe.Sizeof(dataToUpload))
	}
	fmt.Println("File path: ", configStorj.UploadPath)

	fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
	// Uploading BSON to Storj
	err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, &opts)
    if err != nil {
        return fmt.Errorf("Could not upload: %s", err)
	}

	fmt.Println("\nUploading of the object to the Storj bucket: Completed!")
	

	if gb_DEBUG {
		// test uploaded data by downloading it
		serializedAccess, err := access.Serialize()
		
		err = downloadObjectFromStorj(fullFileName, configStorj.UploadPath, serializedAccess)
		if err!=nil{
			return fmt.Errorf("Could not download data: %s", err)
		}		
		
	}
	

	if gb_DEBUG{
		/*
		var listOptions storj.ListOptions
		listOptions.Direction = 127

		listofobject,err:= bucket.ListObjects(ctx, &listOptions)
		fmt.Println("\nListing object from the bucket:",err)
		if err != nil {
			return fmt.Errorf("Listing object failed. %s: ", err)
		}
		fmt.Println("List of object" ,listofobject)
		*/
	}
	
    return nil
}


func downloadObjectFromStorj(fullFileName string, uploadPath string, serializedAccess string) (err error) {
	configStorj, err := loadStorjConfiguration(fullFileName)
	if err != nil {
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}
	
	
	fmt.Println("\nCreating New Uplink...")
    var cfg uplink.Config
	ctx := context.Background()

    // First, create an Uplink handle.
    ul, err := uplink.NewUplink(ctx, &cfg)
    if err != nil {
        return err
    }
    defer ul.Close()

    // Parse the API key. API keys are "macaroons" that allow you to create new,
    // restricted API keys.
    key, err := uplink.ParseAPIKey(configStorj.ApiKey)
    if err != nil {
        return err
    }

    // Open the project in question. Projects are identified by a specific
    // Satellite and API key
    p, err := ul.OpenProject(ctx, configStorj.Satellite, key)
    if err != nil {
        return err
    }
    defer p.Close()

    // Parse the encryption context
    access, err := uplink.ParseEncryptionAccess(serializedAccess)
    if err != nil {
        return err
	}
	
	if gb_DEBUG{
		fmt.Println("Download: Access:", *access)
		fmt.Println("Download: Serialize Access:", serializedAccess)
	}

    // Open bucket
    bucket, err := p.OpenBucket(ctx, configStorj.Bucket, access)
    if err != nil {
        return err
    }
    defer bucket.Close()

	fmt.Println("BSON data is being download from: ", uploadPath)
    // Open file
    obj, err := bucket.OpenObject(ctx, uploadPath)
    if err != nil {
        return err
    }
    defer obj.Close()

    // Get a reader for the entire file
    r, err := obj.DownloadRange(ctx, 0, -1)
    if err != nil {
        return err
    }
    defer r.Close()

	fmt.Println("Starting to download the BSON data...")
    // Read the file
    data, err := ioutil.ReadAll(r)
    if err != nil {
        return err
	}
	err = ioutil.WriteFile("downloaddata.bson", data, 0644)
	fmt.Println("\n\nDownloaded data : ", data)

    return nil
}


// create command-line tool
var app = cli.NewApp()

func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to decentralized Storj cloud"
	app.Author = "Satyam Shivam" 
	app.Version = "1.0.4"
}

func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read and parse JSON information about MongoDB instance properties and then fetch ALL its collections\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_connector.json\n\n     example = ./storj_mongodb d ./config/db_property.json\n\n\n",
			Action: func(cliContext *cli.Context) { 

				data, dbname, err := connectMongo_FetchData( cliContext )	
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}

				if gb_DEBUG {
					fmt.Println("Size of fetched data from database :", dbname,unsafe.Sizeof(data));
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
				if len(cliContext.Args()) > 0 {
					fullFileName = cliContext.Args()[0]
				}

				// sample database name
				dbName := "testdb"
				//
				// sample data to be uploaded
				jsonData := "{'testKey': 'testValue'}"
				// Converting JSON data to bson data 
				bsonData, _ := json.Marshal(jsonData)

				err := connectStorj_UploadData(fullFileName, []byte(bsonData), dbName)
				if err != nil{
					fmt.Println("Error while uploading data to the Storj bucket")
				}
			},
		},
		{
			Name : "connect",
			Aliases : []string{"c"},
			Usage : "use this command to connect and transfer ALL collections from a desired MongoDB instance to given Storj Bucket in BSON format\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties in JSON format\n\n       if this filename is not given, then data is read from ./config/db_property.json\n\n       2. fileName [optional] = provide full file name (with complete path), storing Storj configuration in JSON format\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb c ./config/db_property.json ./config/storj_config.json\n\n\n",
			Action : func(cliContext *cli.Context){
				// default Storj configuration file name
				var fullFileName string = "./config/storj_config.json"
				// reading filename from the command line 
				if len(cliContext.Args()) > 1 {
					fullFileName = cliContext.Args()[1]
				}
				//
				// fetching data from the mongodb
				data,dbname,err := connectMongo_FetchData( cliContext )	
				//
				if err != nil {
					log.Fatalf("connectMongo_FetchData: %s", err)
				}
				//
				if gb_DEBUG {
					fmt.Println("Size of fetched data from database: ", dbname,unsafe.Sizeof(data))
				}
				// connecting to storj network for uploading data
				err = connectStorj_UploadData( fullFileName , []byte(data),dbname)
				if err != nil{
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

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("app.Run: %s", err)
	}
}
