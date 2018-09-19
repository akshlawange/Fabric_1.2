#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# This is a collection of bash functions used by different scripts

ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
PEER0_NatixisSec_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/natixissec.example.com/peers/peer0.natixissec.example.com/tls/ca.crt
PEER0_NatixisBank_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/natixisbank.example.com/peers/peer0.natixisbank.example.com/tls/ca.crt 
PEER0_ScoGenSec_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/socgensec.example.com/peers/peer0.socgensec.example.com/tls/ca.crt
PEER0_ScoGenBank_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/socgenbank.example.com/peers/peer0.socgenbank.example.com/tls/ca.crt
PEER0_WellsFargoSec_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/wellsfargosec.example.com/peers/peer0.wellsfargosec.example.com/tls/ca.crt
PEER0_WellsFargoBank_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/wellsfargobank.example.com/peers/peer0.wellsfargobank.example.com/tls/ca.crt
PEER0_JPMAssetManagement_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/jpmassetmanagement.example.com/peers/peer0.jpmassetmanagement.example.com/tls/ca.crt
PEER0_BroadridgeGlobal_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/broadridgeglobal.example.com/peers/peer0.broadridgeglobal.example.com/tls/ca.crt

# verify the result of the end-to-end test
verifyResult() {
  if [ $1 -ne 0 ]; then
    echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
    echo
    exit 1
  fi
}

# Set OrdererOrg.Admin globals
setOrdererGlobals() {
  CORE_PEER_LOCALMSPID="BroadridgeMSP"
  CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem
  CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/users/Admin@example.com/msp
}


setGlobals () {
  PEER=$1
  ORG=$2

  if [ $ORG -eq 1 ] ; then
     echo "***Setting Variables for NatixisBank:"
     CORE_PEER_LOCALMSPID="NatixisBankMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_NatixisBank_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/natixisbank.example.com/users/Admin@natixisbank.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.natixisbank.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.natixisbank.example.com:7051
     fi
  elif [ $ORG -eq 2 ] ; then
     echo "***Setting Variables for NatixisSec:"
     CORE_PEER_LOCALMSPID="NatixisSecMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_NatixisSec_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/natixissec.example.com/users/Admin@natixissec.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.natixissec.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.natixissec.example.com:7051
     fi
  elif [ $ORG -eq 3 ] ; then
     echo "***Setting Variables for SocGenBank:"
     CORE_PEER_LOCALMSPID="SocGenBankMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ScoGenBank_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/socgenbank.example.com/users/Admin@socgenbank.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.socgenbank.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.socgenbank.example.com:7051
     fi
  elif [ $ORG -eq 4 ] ; then
     echo "***Setting Variables for SocGenSec:"
     CORE_PEER_LOCALMSPID="SocGenSecMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_ScoGenSec_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/socgensec.example.com/users/Admin@socgensec.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.socgensec.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.socgensec.example.com:7051
     fi
  elif [ $ORG -eq 5 ] ; then
     echo "***Setting Variables for WellsFargoBank:"
     CORE_PEER_LOCALMSPID="WellsFargoBankMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_WellsFargoBank_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/wellsfargobank.example.com/users/Admin@wellsfargobank.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.wellsfargobank.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.wellsfargobank.example.com:7051
     fi
  elif [ $ORG -eq 6 ] ; then
     echo "***Setting Variables for WellsFargoSec:"
     CORE_PEER_LOCALMSPID="WellsFargoSecMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_WellsFargoSec_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/wellsfargosec.example.com/users/Admin@wellsfargosec.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.wellsfargosec.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.wellsfargosec.example.com:7051
     fi
  elif [ $ORG -eq 7 ] ; then
     echo "***Setting Variables for JPMAssetManagement:"
     CORE_PEER_LOCALMSPID="JPMAssetManagementMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_JPMAssetManagement_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/jpmassetmanagement.example.com/users/Admin@jpmassetmanagement.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.jpmassetmanagement.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.jpmassetmanagement.example.com:7051
     fi
  elif [ $ORG -eq 8 ] ; then
     echo "***Setting Variables for BroadridgeGlobal:"
     CORE_PEER_LOCALMSPID="BroadridgeGlobalMSP"
     CORE_PEER_TLS_ROOTCERT_FILE=$PEER0_BroadridgeGlobal_CA
     CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/broadridgeglobal.example.com/users/Admin@broadridgeglobal.example.com/msp
     if [ $PEER -eq 0 ]; then
        CORE_PEER_ADDRESS=peer0.broadridgeglobal.example.com:7051
     else
        CORE_PEER_ADDRESS=peer1.broadridgeglobal.example.com:7051
     fi
  fi

  env |grep CORE
}

updateAnchorPeers() {
  PEER=$1
  ORG=$2
  CHANNEL_NAME=$3
  setGlobals $PEER $ORG

  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer channel update -o orderer0.example.com:7050 -c ${CHANNEL_NAME} -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors_${CHANNEL_NAME}.tx >&log.txt
    res=$?
    set +x
  else
    set -x
    peer channel update -o orderer0.example.com:7050 -c ${CHANNEL_NAME} -f ./channel-artifacts/${CORE_PEER_LOCALMSPID}anchors_${CHANNEL_NAME}.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
    res=$?
    set +x
  fi
  cat log.txt
  verifyResult $res "Anchor peer update failed"
  echo "===================== Anchor peers updated for org '$CORE_PEER_LOCALMSPID' on channel '${CHANNEL_NAME}' ===================== "
  sleep $DELAY
  echo
}

## Sometimes Join takes time hence RETRY at least 5 times
joinChannelWithRetry() {
  PEER=$1
  ORG=$2
  CHANNEL_NAME=$3
  setGlobals $PEER $ORG

  set -x
  peer channel join -b ${CHANNEL_NAME}.block >&log.txt
  res=$?
  set +x
  cat log.txt
  if [ $res -ne 0 -a $COUNTER -lt $MAX_RETRY ]; then
    COUNTER=$(expr $COUNTER + 1)
    echo "peer${PEER}.org${ORG} failed to join the channel, Retry after $DELAY seconds"
    sleep $DELAY
    joinChannelWithRetry $PEER $ORG
  else
    COUNTER=1
  fi
  verifyResult $res "After $MAX_RETRY attempts, peer${PEER}.org${ORG} has failed to join channel '${CHANNEL_NAME}' "
}

installChaincode() {
  PEER=$1
  ORG=$2
  CHAINCODE_NAME=$3
  CC_SRC_PATH=$4
  setGlobals $PEER $ORG
  VERSION=${5:-1.0}
  set -x
  peer chaincode install -n ${CHAINCODE_NAME} -v ${VERSION} -l ${LANGUAGE} -p ${CC_SRC_PATH} >&log.txt
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode ${CHAINCODE_NAME} installation on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is installed on peer${PEER}.org${ORG} ===================== "
  echo
}

instantiateChaincode() {
  PEER=$1
  ORG=$2
  CHAINCODE_NAME=$3
  CHANNEL_NAME=$4
  POLICY=$5
  COLLECTION=$6
  setGlobals $PEER $ORG
  VERSION=${5:-1.0}

  # while 'peer chaincode' command can get the orderer endpoint from the peer
  # (if join was successful), let's supply it directly as we know it using
  # the "-o" option
  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer chaincode instantiate -o orderer0.example.com:7050 -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -l ${LANGUAGE} -v ${VERSION} -c '{"Args":[]}' -P ${POLICY} --collections-config ${COLLECTION} >&log.txt
Note
    res=$?
    set +x
  else
    set -x
    peer chaincode instantiate -o orderer0.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -l ${LANGUAGE} -v 1.0 -c '{"Args":[]}' -P ${POLICY} --collections-config ${COLLECTION}  >&log.txt
    res=$?
    set +x
  fi
  cat log.txt
  verifyResult $res "Chaincode instantiation on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' failed"
  echo "===================== Chaincode is instantiated on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  echo
}

upgradeChaincode() {
  PEER=$1
  ORG=$2
  CHAINCODE_NAME=$3
  CHANNEL_NAME=$4
  setGlobals $PEER $ORG

  set -x
  peer chaincode upgrade -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C ${CHANNEL_NAME} -n ${CHAINCODE_NAME} -v 2.0 -c '{"Args":["init","a","90","b","210"]}' -P "AND ('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')"
  res=$?
  set +x
  cat log.txt
  verifyResult $res "Chaincode upgrade on peer${PEER}.org${ORG} has failed"
  echo "===================== Chaincode is upgraded on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' ===================== "
  echo
}


# fetchChannelConfig <channel_id> <output_json>
# Writes the current channel config for a given channel to a JSON file
fetchChannelConfig() {
  CHANNEL=$1
  OUTPUT=$2

  setOrdererGlobals

  echo "Fetching the most recent configuration block for the channel"
  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer channel fetch config config_block.pb -o orderer.example.com:7050 -c $CHANNEL --cafile $ORDERER_CA
    set +x
  else
    set -x
    peer channel fetch config config_block.pb -o orderer.example.com:7050 -c $CHANNEL --tls --cafile $ORDERER_CA
    set +x
  fi

  echo "Decoding config block to JSON and isolating config to ${OUTPUT}"
  set -x
  configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config >"${OUTPUT}"
  set +x
}

# signConfigtxAsPeerOrg <org> <configtx.pb>
# Set the peerOrg admin of an org and signing the config update
signConfigtxAsPeerOrg() {
  PEERORG=$1
  TX=$2
  setGlobals 0 $PEERORG
  set -x
  peer channel signconfigtx -f "${TX}"
  set +x
}

# createConfigUpdate <channel_id> <original_config.json> <modified_config.json> <output.pb>
# Takes an original and modified config, and produces the config update tx
# which transitions between the two
createConfigUpdate() {
  CHANNEL=$1
  ORIGINAL=$2
  MODIFIED=$3
  OUTPUT=$4

  set -x
  configtxlator proto_encode --input "${ORIGINAL}" --type common.Config >original_config.pb
  configtxlator proto_encode --input "${MODIFIED}" --type common.Config >modified_config.pb
  configtxlator compute_update --channel_id "${CHANNEL}" --original original_config.pb --updated modified_config.pb >config_update.pb
  configtxlator proto_decode --input config_update.pb --type common.ConfigUpdate >config_update.json
  echo '{"payload":{"header":{"channel_header":{"channel_id":"'$CHANNEL'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . >config_update_in_envelope.json
  configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope >"${OUTPUT}"
  set +x
}

# parsePeerConnectionParameters $@
# Helper function that takes the parameters from a chaincode operation
# (e.g. invoke, query, instantiate) and checks for an even number of
# peers and associated org, then sets $PEER_CONN_PARMS and $PEERS
parsePeerConnectionParameters() {
  # check for uneven number of peer and org parameters
  if [ $(($# % 2)) -ne 0 ]; then
    exit 1
  fi

  PEER_CONN_PARMS=""
  PEERS=""
  while [ "$#" -gt 0 ]; do
    PEER="peer$1.org$2"
    PEERS="$PEERS $PEER"
    PEER_CONN_PARMS="$PEER_CONN_PARMS --peerAddresses $PEER.example.com:7051"
    if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "true" ]; then
      TLSINFO=$(eval echo "--tlsRootCertFiles \$PEER$1_ORG$2_CA")
      PEER_CONN_PARMS="$PEER_CONN_PARMS $TLSINFO"
    fi
    # shift by two to get the next pair of peer/org parameters
    shift
    shift
  done
  # remove leading space for output
  PEERS="$(echo -e "$PEERS" | sed -e 's/^[[:space:]]*//')"
}

# chaincodeInvoke <peer> <org> ...
# Accepts as many peer/org pairs as desired and requests endorsement from each
chaincodeInvoke() {
  parsePeerConnectionParameters $@
  res=$?
  verifyResult $res "Invoke transaction failed on channel '$CHANNEL_NAME' due to uneven number of peer and org parameters "

  # while 'peer chaincode' command can get the orderer endpoint from the
  # peer (if join was successful), let's supply it directly as we know
  # it using the "-o" option
  if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
    set -x
    peer chaincode invoke -o orderer.example.com:7050 -C $CHANNEL_NAME -n mycc $PEER_CONN_PARMS -c '{"Args":["invoke","a","b","10"]}' >&log.txt
    res=$?
    set +x
  else
    set -x
    peer chaincode invoke -o orderer.example.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n mycc $PEER_CONN_PARMS -c '{"Args":["invoke","a","b","10"]}' >&log.txt
    res=$?
    set +x
  fi
  cat log.txt
  verifyResult $res "Invoke execution on $PEERS failed "
  echo "===================== Invoke transaction successful on $PEERS on channel '$CHANNEL_NAME' ===================== "
  echo
}
