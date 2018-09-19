#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Build your Distributed Ledger Repo end-to-end test"
echo
GLOBAL_CHANNEL_NAME="$1"
REPO_CHANNEL_NAME="$2"
DELAY="$3"
LANGUAGE="$4"
TIMEOUT="$5"
VERBOSE="$6"
: ${CHANNEL_NAME:="mychannel"}
: ${DELAY:="3"}
: ${LANGUAGE:="golang"}
: ${TIMEOUT:="10"}
: ${VERBOSE:="false"}
LANGUAGE=`echo "$LANGUAGE" | tr [:upper:] [:lower:]`
COUNTER=1
MAX_RETRY=5

#CC_SRC_PATH="github.com/chaincode/chaincode_example02/go/"
ASSETOWNERSHIP_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/AssetOwnershipPrivateCC" 
ASSETTOKEN_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/AssetTokenPrivateCC" 
REPODEAL_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/RepoDealPrivateCC" 
MULTIPARTY_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/MultiPartyPrivateCC" 
SETTLEMENT_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/SettlementPrivateCC" 
VALUATION_CC_SRC_PATH="github.com/hyperledger/fabric/examples/chaincode/go/ValuationPrivateCC"
COLLECTION_PATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/scripts/"

if [ "$LANGUAGE" = "node" ]; then
	CC_SRC_PATH="/opt/gopath/src/github.com/chaincode/chaincode_example02/node/"
fi

echo "Global Channel name : "$GLOBAL_CHANNEL_NAME
echo "Repo Channel name : "$REPO_CHANNEL_NAME

# import utils
. scripts/utils.sh

createGlobalChannel() {
        setGlobals 8 0 

        if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                peer channel create -o orderer0.example.com:7050 -c $GLOBAL_CHANNEL_NAME -f ./channel-artifacts/globalchannel.tx >&log.txt
        else
                peer channel create -o orderer0.example.com:7050 -c $GLOBAL_CHANNEL_NAME -f ./channel-artifacts/globalchannel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
        fi
        res=$?
        cat log.txt
        verifyResult $res "Global Channel creation failed"
        echo "===================== Channel \"$GLOBAL_CHANNEL_NAME\" is created successfully ===================== "
        echo
}

createRepoChannel() {
        setGlobals 8 0

        if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                peer channel create -o orderer0.example.com:7050 -c $REPO_CHANNEL_NAME -f ./channel-artifacts/repochannel.tx >&log.txt
        else
                peer channel create -o orderer0.example.com:7050 -c $REPO_CHANNEL_NAME -f ./channel-artifacts/repochannel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
        fi
        res=$?
        cat log.txt
        verifyResult $res "Repo Channel creation failed"
        echo "===================== Channel \"$REPO_CHANNEL_NAME\" is created successfully ===================== "
        echo
}

joinChannel () {
#        #for org in 1 2 3 4 5 6 7 8; do
#	for org in 1 2; do
#	    for peer in 0 1; do
#		joinChannelWithRetry $peer $org $GLOBAL_CHANNEL_NAME
#		echo "===================== peer${peer}.org${org} joined channel '$GLOBAL_CHANNEL_NAME' ===================== "
#		sleep $DELAY
#		echo
#	    done
#	done

	for org in 1 2 3 4 5 6 7 8; do
	    for peer in 0 1; do
		joinChannelWithRetry $peer $org $REPO_CHANNEL_NAME
		echo "===================== peer${peer}.org${org} joined channel '$REPO_CHANNEL_NAME' ===================== "
		sleep $DELAY
		echo
	    done
	done
}

updateAnchor () {
#        #for org in 1 2 3 4 5 6 7 8; do
#        for org in 1 2; do
#            for peer in 0; do
#                echo "===================== Updating anchor peers for '$GLOBAL_CHANNEL_NAME', org '$org' and peer '$peer'  ===================== "
#                updateAnchorPeers $peer $org $GLOBAL_CHANNEL_NAME
#                sleep $DELAY
#                echo
#            done
#        done

        for org in 1 2 3 4 5 6 7 8; do
            for peer in 0; do
                echo "===================== Updating anchor peers for '$REPO_CHANNEL_NAME', org '$org' and peer '$peer'  ===================== "
                updateAnchorPeers $peer $org $REPO_CHANNEL_NAME
                sleep $DELAY
                echo
            done
        done

}

installAssetOwnershipPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2;  do
            for peer in 0 1; do
                echo "===================== Installing AssetOwnershipPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org AssetOwnershipPrivateCC $ASSETOWNERSHIP_CC_SRC_PATH
                echo "===================== Installation Completed for AssetOwnershipPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

installAssetTokenPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2; do
            for peer in 0 1; do
                echo "===================== Installing AssetTokenPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org AssetTokenPrivateCC $ASSETTOKEN_CC_SRC_PATH
                echo "===================== Installation Completed for AssetTokenPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

installSettlementPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2; do
            for peer in 0 1; do
                echo "===================== Installing SettlementPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org SettlementPrivateCC $SETTLEMENT_CC_SRC_PATH
                echo "===================== Installation Completed for SettlementPrivateCC Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

installRepoDealPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2; do
            for peer in 0 1; do
                echo "===================== Installing RepoDealPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org RepoDealPrivateCC $REPODEAL_CC_SRC_PATH
                echo "===================== Installation Completed for RepoDealPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

installMultiPartyPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2; do
            for peer in 0 1; do
                echo "===================== Installing MultiPartyPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org MultiPartyPrivateCC $MULTIPARTY_CC_SRC_PATH
                echo "===================== Installation Completed for MultiPartyPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

installValuationPrivateChaincode () {
        #for org in 1 2 3 4 5 6 7 8; do
        for org in 1 2; do
            for peer in 0 1; do
                echo "===================== Installing ValuationPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                installChaincode $peer $org ValuationPrivateCC $VALUATION_CC_SRC_PATH
                echo "===================== Installation Completed for ValuationPrivate Chaincode for org '$org' and peer '$peer'  ===================== "
                sleep $DELAY
                echo
            done
        done
}

## Sleep 10 secs
sleep 10

## Create channel
#echo "Creating Global channel..."
#createGlobalChannel

echo "Creating Repo channel..."
createRepoChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel

## Set the anchor peers for each org in the channel
echo "Updating all peers Anchor..."
updateAnchor

#echo "Install AssetOwnership Chaincode..."
#installAssetOwnershipPrivateChaincode

#echo "Install AssetToken Chaincode..."
#installAssetTokenPrivateChaincode


#echo "Install Repo Chaincode..."
installRepoDealPrivateChaincode
#echo "Install MultiParty Chaincode..."
installMultiPartyPrivateChaincode
#echo "Install Valuation Chaincode..."
installValuationPrivateChaincode

#echo "Install Settlement Chaincode..."
#installSettlementPrivateChaincode

#echo "Instantiating AssetOwnership chaincode..."
#instantiateChaincode 0 1 AssetOwnershipPrivateCC globalchannel "AND('NatixisBankMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetOwnershipPrivateCC_NatixisBank.json
#echo "Instantiating AssetToken chaincode..."
#instantiateChaincode 0 1 AssetTokenPrivateCC globalchannel "AND('NatixisBankMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetTokenPrivateCC_NatixisBank.json

#echo "Instantiating AssetOwnership chaincode..."
#instantiateChaincode 0 2 AssetOwnershipPrivateCC globalchannel "AND('NatixisSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetOwnershipPrivateCC_NatixisSec.json
#echo "Instantiating AssetToken chaincode..."
#instantiateChaincode 0 2 AssetTokenPrivateCC globalchannel "AND('NatixisSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetTokenPrivateCC_NatixisSec.json

#echo "Instantiating AssetOwnership chaincode..."
#instantiateChaincode 0 2 AssetOwnershipPrivateCC globalchannel "AND('SocGenBankMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetOwnershipPrivateCC_SocGenBank.json
#echo "Instantiating AssetToken chaincode..."
#instantiateChaincode 0 2 AssetTokenPrivateCC globalchannel "AND('SocGenBankMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetTokenPrivateCC_SocGenBank.json

#echo "Instantiating AssetOwnership chaincode..."
#instantiateChaincode 0 2 AssetOwnershipPrivateCC globalchannel "AND('SocGenSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetOwnershipPrivateCC_SocGenSec.json
#echo "Instantiating AssetToken chaincode..."
#instantiateChaincode 0 2 AssetTokenPrivateCC globalchannel "AND('SocGenSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/AssetTokenPrivateCC_SocGenSec.json

echo "Instantiating Repo chaincode..."
instantiateChaincode 0 1 RepoDealPrivateCC repochannel "AND('NatixisBankMSP.member','NatixisSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/RepoDealPrivateCC_NatixisBankSec.json
echo "Instantiating MultiParty chaincode..."
instantiateChaincode 0 1 MultiPartyPrivateCC repochannel "AND('NatixisBankMSP.member','NatixisSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/MultiPartyPrivateCC_NatixisBankSec.json
echo "Instantiating Valuation chaincode..."
instantiateChaincode 0 1 ValuationPrivateCC repochannel "AND('NatixisBankMSP.member','NatixisSecMSP.member','BroadridgeGlobalMSP.member')" $COLLECTION_PATH/ValuationPrivateCC_NatixisBankSec.json


echo
echo "========= All GOOD, BYFN execution completed =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
