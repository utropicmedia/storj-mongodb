# create module path, if it does not exist
mkdir -p $HOME/go/src/utropicmedia/mongodb_storj_interface/mongo
mkdir -p $HOME/go/src/utropicmedia/mongodb_storj_interface/storj
# move go packages to include path
#rm -rf $HOME/go/src/mongo
#rm -rf $HOME/go/src/storj
cp utropicmedia/mongodb_storj_interface/mongo/* $HOME/go/src/utropicmedia/mongodb_storj_interface/mongo/
cp utropicmedia/mongodb_storj_interface/storj/* $HOME/go/src/utropicmedia/mongodb_storj_interface/storj/
cp utropicmedia/mongodb_storj_interface/* $HOME/go/src/utropicmedia/mongodb_storj_interface/
