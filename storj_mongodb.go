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
    "os"
    "log"
    "fmt"
    "time"    
    "bufio"
    "context"
    "io/ioutil"
    "encoding/json"
    "github.com/urfave/cli"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
    "storj.io/storj/lib/uplink"
    "storj.io/storj/pkg/storj"
    "bytes"
    "unsafe"


)

var gb_DEBUG bool
var apikey string =""
var satellite string =""
var bucket string =""
var uploadPath string =""
var encryptionPassphrase string =""
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
	Encryptionpassphrase string `json:"encryptionpassphrase"`
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

func storjConnect(cliContext *cli.Context)(error){
	
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
	apiKeyparse, err := uplink.ParseAPIKey(configStorj.ApiKey)
	fmt.Println("API key ",apiKeyparse)
	if err != nil {
		log.Fatalln("could not parse api key:", err)
	}

	data :="{'test':'testdata'}"
	  json2BSON, _ := json.Marshal(data)

  // Debug
  fmt.Println("BSON :\n",json2BSON)
  var bson2JSON interface{}
  err = json.Unmarshal(json2BSON, &bson2JSON)

  if err != nil {
   fmt.Println(err)
  }

  //Debug
  fmt.Println("JSON : \n", bson2JSON)
	//data,err:=connectMongo(cliContext)	
	if err != nil {
		log.Fatalf("readPropertiesDB: %s", err)
	}
	fmt.Println("Error ",unsafe.Sizeof(data))
	err = UploadDataToStorj(context.Background(),configStorj.Satellite, configStorj.Encryptionpassphrase, apiKeyparse, configStorj.Bucket, configStorj.UploadPath, []byte(data))
	return (nil)
	
}

func connectMongo(cliContext *cli.Context)([]byte,error){
	
	configMongoDB, err := readPropertiesDB(cliContext)
	var alldata =[]byte{}			//
	if err != nil {
		log.Fatalf("readPropertiesDB: %s", err)
	}
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
	    return alldata,nil
	}


	// apiKeyparse, err := uplink.ParseAPIKey(apikey)
	// 	if err != nil {
	// 	log.Fatalln("could not parse api key:", err)
	// 	}
	
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
			//fmt.Println("\nBSON", rawDocumentBSON)
			//alldata =alldata+rawDocumentBSON
			alldata =append(alldata[:],rawDocumentBSON...)
			//err=UploadDataToStorj(context.Background(),satellite, encryptionPassphrase, apiKeyparse, bucket, uploadPath, []byte(rawDocumentBSON),collectionName)
			if err != nil {
				log.Fatalln("error:", err)
			}

			if gb_DEBUG {
				// DE-BUG
				var bson2JSON interface{}
				err = bson.Unmarshal(rawDocumentBSON, &bson2JSON)
				//fmt.Println("JSON equivalent: ", bson2JSON)
				//
				if err != nil {
					fmt.Println(err)
				}
			}
			//break;
	    }
	    //
	    //err = UploadDataToStorj(context.Background(),satellite, encryptionPassphrase, apiKeyparse, bucket, uploadPath, []byte(alldata),collectionName)
	    //err = UploadDataToStorj(context.Background(),satellite, encryptionPassphrase+collectionName, apiKeyparse, bucket, uploadPath, []byte(alldata))
			
	}
	//fmt.Println(alldata)
	return alldata,nil
}
// create command-line tool
var app = cli.NewApp()

func setAppInfo() {
	app.Name = "Storj MongoDB Connector"
	app.Usage = "Backup your MongoDB collections to decentralized Storj cloud"
	app.Author = "Satyam Shivam" 
	app.Version = "1.0.2"
}

func UploadDataToStorj(ctx context.Context,satelliteAddress string, encryptionPassphrase string, apiKey uplink.APIKey,
    bucketName, uploadPath string, dataToUpload []byte) error {
    //ctx = context.Background()

    //upl, err := uplink.NewUplink(ctx, nil)
    fmt.Println("Creating New Uplink Object")
    var cfg uplink.Config
    cfg.Volatile.TLS.SkipPeerCAWhitelist = true
    upl, err := uplink.NewUplink(ctx, &cfg)
    //upl, err := uplink.NewUplink(ctx, nil)
    if err != nil {
        return fmt.Errorf("could not create new Uplink object: %v", err)
    }
    defer upl.Close()


    fmt.Println("Open Project ")
    proj, err := upl.OpenProject(ctx, satelliteAddress, apiKey)
    fmt.Println("Open Project err",err)
    fmt.Println("Open Project err",proj)
    if err != nil {
        return fmt.Errorf("could not open project: %v", err)
    }
    defer proj.Close()

    

    fmt.Println("Create Bucket " ,bucketName)
    _, err = proj.CreateBucket(ctx, bucketName, nil)
    fmt.Println("Create Bucket " ,err)
    if err != nil {
        return fmt.Errorf("could not create bucket: %v", err)
    }

    fmt.Println("bucket info")
    bucketinfo,cfgbucket,err:=proj.GetBucketInfo(ctx,bucketName)
    fmt.Println("Error  " ,err)
    fmt.Println("Bucket info " ,bucketinfo)
    if err != nil {
        return fmt.Errorf("could not get info: %v", err)
    }
    defer proj.Close()
    
    fmt.Println("CFG bucket info ",cfgbucket.Volatile)
    fmt.Println("CFG bucket info ",bucketinfo)


	
    //fmt.Println("Create Bucket cfg" ,cfg)


    fmt.Println("Creating new encryption ",encryptionPassphrase)

    fmt.Println("List of bucket ")
    list := uplink.BucketListOptions{
		Direction: storj.Forward}
	for {
		result, err := proj.ListBuckets(ctx, &list)
		if err != nil {
			return err
		}
		for _, bucket := range result.Items {
			fmt.Println("Bucket: %v\n", bucket.Name)
		}
		if !result.More {
			break
		}
		list = list.NextPage(result)
	}


    fmt.Println("Phrase ")
    encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, encryptionPassphrase)
    fmt.Println("Phra errpr",err)
    if err != nil {
        return fmt.Errorf("could not create encryption key: %v", err)
    }
    access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)


    fmt.Println("Opening created bucket",bucketName)
    // Open up the desired Bucket within the Project
    
    fmt.Println("Access ",access)
    bucket, err := proj.OpenBucket(ctx, bucketName, access)
    fmt.Println("Opening bucket",err)
    if err != nil {
        return fmt.Errorf("could not open bucket %q: %v", bucketName, err)
    }
    defer bucket.Close()

    //BucketListOptions := storj.BucketListOptions

    //var bucketListOptions = storj.BucketListOptions
    listofobject,err:= bucket.ListObjects(ctx, nil)
    fmt.Println("Listing obejct from the bucket",err)
    if err != nil {
        return fmt.Errorf("Listing object failed. %q: %v", err)
    }
    fmt.Println("List of object" ,listofobject)
 //    fmt.Println("Converting Data to Buffer");
 	buf := bytes.NewBuffer(dataToUpload)



    //fmt.Println("BSON DATA" ,dataToUpload);

    //var bson2JSON interface{}
	//err = bson.Unmarshal(dataToUpload, &bson2JSON)
	//fmt.Println("JSON equivalent: ", bson2JSON)
	//
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("\n\nBuffer data")
	fmt.Println(unsafe.Sizeof(buf))
	
    //fmt.Println("Buffer Data" ,buf);

    t := time.Now()
    time :=t.Format("2006-01-02_15:04:05.000000")
    var filename string = time+".bson"
    uploadPath =uploadPath+filename


    opts :=&uplink.UploadOptions{}
    opts.Volatile.RedundancyScheme = cfgbucket.Volatile.RedundancyScheme
	opts.Volatile.EncryptionParameters = cfgbucket.EncryptionParameters

    fmt.Println("\nData Uploaded")
	fmt.Println(unsafe.Sizeof(dataToUpload))
    //uploadData := []byte("HELLO, hOW ARE YOU?")

    //bufTest := bytes.NewBuffer(uploadData)
    err = bucket.UploadObject(ctx, uploadPath, buf, opts)
    fmt.Println("Error uploading object ",err);
    if err != nil {
        return fmt.Errorf("could not upload: %v", err)
    }

    // fmt.Println("Opening bucket for downloading");
	// Initiate a download of the same object again
	//uploadPath="data.bsonbeginnersbook_2019-07-16_12:40:09.528607.txt"
	readBack, err := bucket.OpenObject(ctx, uploadPath)
	fmt.Println("Readback ",readBack)
	fmt.Println("Readback error ",err)
	if err != nil {
		return fmt.Errorf("could not open object at %q: %v", uploadPath, err)
	}
	defer readBack.Close()

	fmt.Println("Downloading range");
	// We want the whole thing, so range from 0 to -1
	strm, err := readBack.DownloadRange(ctx, 0, -1)
	if err != nil {
		return fmt.Errorf("could not initiate download: %v", err)
	}
	defer strm.Close()
	fmt.Println("Read from the stream");
	// Read everything from the stream
	receivedContents, err := ioutil.ReadAll(strm)
	if err != nil {
		return fmt.Errorf("could not read object: %v", err)
	}

	fmt.Println("\n\nRecievd data")
	fmt.Println(unsafe.Sizeof(receivedContents))
	if !bytes.Equal(receivedContents, dataToUpload) {
        return fmt.Errorf("got different object back: %q != %q", dataToUpload, receivedContents)
    }
    
    return nil
}




func setCommands() {

	app.Commands = []cli.Command{
		{
			Name:    "mongodb.props",
			Aliases: []string{"d"},
			Usage:   "use this command to read properties of desired mongo database\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_connector\n\n     example = ./storj_mongodb d ./config/db_property\n\n\n",
			Action: func(cliContext *cli.Context) { 

				// configMongoDB, err := readPropertiesDB(cliContext)
				// //
				// if err != nil {
				// 	log.Fatalf("readPropertiesDB: %s", err)
				// }
				
				// // print to debug
				// fmt.Println("Read MongoDB credentials from the ", configMongoDB.fullFileName, " file")
				// fmt.Println("Host Name	: ", configMongoDB.hostName)
				// fmt.Println("Port Number\t: ", configMongoDB.portNumber)
				// fmt.Println("User Name	: ", configMongoDB.userName)
				// fmt.Println("Password	: ", configMongoDB.password)
				// fmt.Println("Database	: ", configMongoDB.database)
				data,err:=connectMongo(cliContext)	
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				fmt.Println(unsafe.Sizeof(data));
			},
		},
		{
			Name:    "storj.config",
			Aliases: []string{"s"},
			Usage:   "use this command to read and parse JSON information about Storj network\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing Storj configuration information\n\n       if this filename is not given, then data is read from ./config/storj_config.json\n\n     example = ./storj_mongodb s ./config/storj_config.json\n\n\n",
			Action: func(cliContext *cli.Context) { 

				// default full file name
				//var fullFileName string="./config/storj_config.json"

				// if len(cliContext.Args()) > 0 {
				// 	fullFileName = cliContext.Args()[0]
				// }
				result:=storjConnect(cliContext)
				fmt.Println("Data Recieved from the function ",result)
			},
		},
		{
			Name : "connect",
			Aliases : []string{"c"},
			Usage : "use this command to connect with a mongoDB instance\n\n     arguments-\n\n       1. fileName [optional] = provide full file name (with complete path), storing mongoDB properties\n\n       if this filename is not given, then data is read from ./config/db_property\n\n     example = ./storj_mongodb c ./config/db_property\n\n\n",
			Action : func(cliContext *cli.Context){
				// read MongoDB's properties, credentials, and database name from a file
				//configMongoDB, err := readPropertiesDB(cliContext)
				//
				// if err != nil {
				// 	log.Fatalf("readPropertiesDB: %s", err)
				// }
				var fullFileName string="./config/storj_config.json"
				if len(cliContext.Args()) > 1 {
					fullFileName = cliContext.Args()[1]
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
				apiKeyparse, err := uplink.ParseAPIKey(configStorj.ApiKey)
				fmt.Println("API key ",apiKeyparse)
				if err != nil {
					log.Fatalln("could not parse api key:", err)
				}
				
				data,err:=connectMongo(cliContext)	
				if err != nil {
					log.Fatalf("readPropertiesDB: %s", err)
				}
				fmt.Println("Data fetched ",unsafe.Sizeof(data))
				err = UploadDataToStorj(context.Background(),configStorj.Satellite, encryptionPassphrase, apiKeyparse, configStorj.Bucket, configStorj.UploadPath, []byte(data))
				// return (nil)
				//fmt.Println("Error ",unsafe.Sizeof(data))
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
