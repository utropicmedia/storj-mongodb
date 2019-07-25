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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	
	"storj.io/storj/lib/uplink"
)


var DEBUG bool = true
var gb_DEBUG_DEV bool = true


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
		/*_, err = proj.CreateBucket(ctx, configStorj.Bucket, nil)
	    if err != nil {
	        return fmt.Errorf("could not create bucket: %v", err)
	    }
		defer proj.Close()*/
	}

	// Creating an encryption key from encryption passphrase
    fmt.Println("\nGet encryption key from pass phrase...")
	
	encryptionKey, err := proj.SaltedKeyFromPassphrase(ctx, configStorj.EncryptionPassphrase)
    if err != nil {
        return fmt.Errorf("Could not create encryption key: %s", err)
	}
 	

	// Creating an encryption context
	access := uplink.NewEncryptionAccessWithDefaultKey(*encryptionKey)
	fmt.Println("\nEncryption access \t:",*access)

	// Serializing the parsed access, so as to compare with the original key
	serializedAccess, err := access.Serialize()
	if err != nil {
		fmt.Println("Error Serialized key : ", err)	
	}
	//
	fmt.Println("Serialized access key\t:", serializedAccess)
	
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

	
	fmt.Println("\nCreating file name in the bucket, as per current time...")
	t := time.Now()
    time := t.Format("2006-01-02_15:04:05")
    var filename string = databaseName + "_" + time + ".bson"
    configStorj.UploadPath = configStorj.UploadPath + filename
    
	fmt.Println("File path: ", configStorj.UploadPath)

	fmt.Println("\nUploading of the object to the Storj bucket: Initiated...")
	// Uploading BSON to Storj
	err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, nil)
    if err != nil {
    	fmt.Println("Uploading of data failed :\n %s",err)
    	fmt.Println("\nRetrying to Uploading data .....\n")
        err = bucket.UploadObject(ctx, configStorj.UploadPath, buf, nil)
        if err != nil {
        	return fmt.Errorf("Could not upload: %s", err)
		}
	}

	fmt.Println("\nUploading of the object to the Storj bucket: Completed!")
	

	if DEBUG {
		// test uploaded data by downloading it
		// serializedAccess, err := access.Serialize()
		// Initiate a download of the same object again
		readBack, err := bucket.OpenObject(ctx, configStorj.UploadPath)
		if err != nil {
			return fmt.Errorf("could not open object at %q: %v", configStorj.UploadPath, err)
		}
		defer readBack.Close()

		fmt.Println("Downloading range")
		// We want the whole thing, so range from 0 to -1
		strm, err := readBack.DownloadRange(ctx, 0, -1)
		if err != nil {
			return fmt.Errorf("could not initiate download: %v", err)
		}
		defer strm.Close()
		fmt.Println("Downloading Object from bucket : Initiated....")
		// Read everything from the stream
		receivedContents, err := ioutil.ReadAll(strm)
		if err != nil {
			return fmt.Errorf("could not read object: %v", err)
		}
		var filenamedownload string = "downloadeddata_" + time + ".bson"
		err = ioutil.WriteFile(filenamedownload, receivedContents, 0644)

		if !bytes.Equal(dataToUpload, receivedContents) {
			return fmt.Errorf("error: uploaded data != downloaded data")
		}
		fmt.Println("Downloading Object from bucket : Complete!")		
	}
	
    return nil
}

