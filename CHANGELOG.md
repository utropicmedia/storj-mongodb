# Storj-MongoDB Changelog

## [1.0.11] - 05-02-2020
### Changelog:
* Changes made accodring to latest uplink v0.31.13.
* Updated uplink, storj and other libraries.
* Simplified code structure.
* Removed unneeded aliases in mongo.go.
* Fixed path handeling in stroj.go.
* Added a downloadObject fucntion in storj.go
* Replaced partnerID with userAgent and named as MongoDB.

## [1.0.10] - 17-12-2019
### Changelog:
* Changess made according to latest libuplink v0.27.1
* Changes made according to updated cli package.
* Added Macroon functionality.
* Added option to access storj using Serialized Scope Key. 
* Added keyword `key` to access storj using API key rather than Serialized Scope Key (defalt).
* Added keyword `restrict` to apply restrictions on API key and provide shareable Serialized Scope Key for users.
* Error handling for various events.