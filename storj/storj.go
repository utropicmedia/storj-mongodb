/*
 * storj package connects to a Storj network,
 * based on the configuration, read from a JSON file.
 * It then stores given BSON data at desired bucket.
 *
 * v 1.0.0
 * Storj functions collected into a separate package 
 */

package storj

import(
	"os"
	"fmt"
	"time"
	"bytes"
	"unsafe"
	"context"
	"reflect"
	"io/ioutil"
	"encoding/json"
	"storj.io/storj/lib/uplink"
	"storj.io/storj/pkg/storj"
)


var DEBUG bool = false
var gb_DEBUG_DEV bool = false


// ConfigStorj depicts keys to search for
// within the stroj_config.json file
type ConfigStorj struct { 
	ApiKey     string `json:"apikey"`
	Satellite  string `json:"satellite"`
	Bucket     string `json:"bucket"`
	UploadPath string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
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

	// Display read information
	fmt.Println("Read Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.ApiKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)
	fmt.Println("Upload Path\t: ", configStorj.UploadPath)

	return configStorj, nil
}


// ConnectStorj_UploadData reads Storj configuration from given file,
// connects to the desired Storj network, and
// uploads given object to the desired bucket
func ConnectStorj_UploadData(fullFileName string, dataToUpload []byte, databaseName string) error {
	
	// Read Storj bucket's configuration from an external file
	configStorj, err := LoadStorjConfiguration(fullFileName)
	if err != nil {
		return fmt.Errorf("loadStorjConfiguration: %s", err)
	}
	
	if DEBUG {
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
	if DEBUG { 
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

	
	if gb_DEBUG_DEV {
		// for creating bucket and listing object in bucket.
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


	}


	if gb_DEBUG_DEV {
		//Bucket info and creating new bucket
		bucketinfo,cfgbucket,err:=proj.GetBucketInfo(ctx,configStorj.Bucket)
		if err != nil {
			// create desired bucket in the project
			fmt.Println("Creating a new bucket: ", configStorj.Bucket)
			//
		    _, err = proj.CreateBucket(ctx, configStorj.Bucket, nil)
		    if err != nil {
		        return fmt.Errorf("could not create bucket: %v", err)
		    }
		}
		defer proj.Close()
		opts :=&uplink.UploadOptions{}
		fmt.Println("Bucket Info : ",bucketinfo)
		opts.Volatile.RedundancyScheme = cfgbucket.Volatile.RedundancyScheme
		opts.Volatile.EncryptionParameters = cfgbucket.EncryptionParameters
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

	if DEBUG {
		// Serializing the parsed access, so as to compare with the original key
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

	if DEBUG {
		fmt.Println("Upload: Type to the data:", reflect.TypeOf(dataToUpload))
		fmt.Println("Upload: Type to the buffer:", reflect.TypeOf(buf))
	}

	
	fmt.Println("\nCreating file name in the bucket, as per current time...")
	t := time.Now()
    time := t.Format("2006-01-02_15:04:05")
    var filename string = databaseName + "_" + time + ".bson"
    configStorj.UploadPath = configStorj.UploadPath + filename


	var opts uplink.UploadOptions
	
    
    if DEBUG {
		fmt.Println("Size of data uploading: ", unsafe.Sizeof(dataToUpload))
	}
	fmt.Println("File path: ", configStorj.UploadPath)

	fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
	// Uploading BSON to Storj
	err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, &opts)
    if err != nil {
    	fmt.Println("Uploading of data failed :\n %s",err)
    	fmt.Println("\nRetrying to Uploading data .....\n")
        err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, &opts)
        if err != nil {
        return fmt.Errorf("Could not upload: %s", err)
		}
	}

	fmt.Println("\nUploading of the object to the Storj bucket: Completed!")
	

	if DEBUG {
		// test uploaded data by downloading it
		serializedAccess, err := access.Serialize()
		
		err = DownloadObjectFromStorj(fullFileName, configStorj.UploadPath, serializedAccess)
		if err != nil{
			return fmt.Errorf("Could not download data: %s", err)
		}		
		
	}
	

	if gb_DEBUG_DEV {
		// download all objects from given bucket and display on the screen
		var listOptions storj.ListOptions
		listOptions.Direction = 127

		listofobject,err:= bucket.ListObjects(ctx, &listOptions)
		fmt.Println("\nListing object from the bucket:", err)
		if err != nil {
			return fmt.Errorf("Listing object failed. %s: ", err)
		}
		fmt.Println("List of object" ,listofobject)
	}
	
    return nil
}


// downloadObjectFromStorj downloads an object, with given serialized key, from the Storj bucket
func DownloadObjectFromStorj(fullFileName string, uploadPath string, serializedAccess string) (err error) {
	configStorj, err := LoadStorjConfiguration(fullFileName)
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
	
	if DEBUG{
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
