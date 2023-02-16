package main

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func main() {
	// Create a new SDK instance using the configuration file
	configProvider := config.FromFile("path/to/config.yaml")
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		panic(fmt.Errorf("failed to create SDK: %v", err))
	}
	defer sdk.Close()

	// Create a new channel client instance
	clientContext := sdk.ChannelContext("mychannel", fabsdk.WithUser("user1"), fabsdk.WithOrg("Org1"))
	channelClient, err := channel.New(clientContext)
	if err != nil {
		panic(fmt.Errorf("failed to create channel client: %v", err))
	}

	// Query the ledger using the channel client
	response, err := channelClient.Query(channel.Request{ChaincodeID: "mychaincode", Fcn: "query", Args: [][]byte{[]byte("key")}})
	if err != nil {
		panic(fmt.Errorf("failed to query ledger: %v", err))
	}

	fmt.Printf("Query result: %s\n", string(response.Payload))
}
