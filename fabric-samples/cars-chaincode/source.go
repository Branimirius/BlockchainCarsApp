package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type CarMalfunction struct {
	Description string
	RepairPrice float32
}

type CarAsset struct {
	ID              string
	Brand           string
	Model           string
	Year            int
	Color           string
	OwnerID         string
	Price           float32
	MalfunctionList []CarMalfunction
}

type PersonAsset struct {
	ID                 string
	FirstName          string
	LastName           string
	EmailAddress       string
	AmountOfMoneyOwned float32
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	carAssets := []CarAsset{
		{ID: "car1", Brand: "Opel", Model: "Cascada", Year: 2013, Color: "blue", OwnerID: "person1", Price: 2500, MalfunctionList: []CarMalfunction{
			{Description: "Shakey steering wheel", RepairPrice: 50},
			{Description: "Oil leaking", RepairPrice: 75},
		}},
		{ID: "car2", Brand: "Audi", Model: "A4", Year: 2016, Color: "red", OwnerID: "person2", Price: 5000, MalfunctionList: []CarMalfunction{
			{Description: "Flat fron left tyre", RepairPrice: 15},
		}},
		{ID: "car3", Brand: "Volvo", Model: "V60", Year: 2014, Color: "green", OwnerID: "person1", Price: 3400, MalfunctionList: []CarMalfunction{
			{Description: "Cracked windscreen", RepairPrice: 100},
			{Description: "Loose back wiper", RepairPrice: 5},
		}},
		{ID: "car4", Brand: "Zastava", Model: "Yugo 45", Year: 1985, Color: "yellow", OwnerID: "person1", Price: 200, MalfunctionList: []CarMalfunction{
			{Description: "Broken alternator", RepairPrice: 50},
			{Description: "Broken spark plug", RepairPrice: 30},
			{Description: "Loose exhaust pipe", RepairPrice: 10},
			{Description: "Overheating", RepairPrice: 80},
		}},
		{ID: "car5", Brand: "Mercedes-Benz", Model: "A-class", Year: 2018, Color: "black", OwnerID: "person3", Price: 5300, MalfunctionList: []CarMalfunction{}},
		{ID: "car6", Brand: "BMW", Model: "X5", Year: 2018, Color: "white", OwnerID: "person2", Price: 6000, MalfunctionList: []CarMalfunction{
			{Description: "Cracked headlight", RepairPrice: 30},
		}},
	}

	personAssets := []PersonAsset{
		{ID: "person1", FirstName: "Branimir", LastName: "Maricic", EmailAddress: "branimir@gmail.com", AmountOfMoneyOwned: 5100.74},
		{ID: "person2", FirstName: "Filip", LastName: "Pinjuh", EmailAddress: "filip@gmail.com", AmountOfMoneyOwned: 3200.0},
		{ID: "person3", FirstName: "Milovan", LastName: "Jovanovic", EmailAddress: "milovan@gmail.com", AmountOfMoneyOwned: 430.0},
	}

	for _, carAsset := range carAssets {
		carAssetJSON, err := json.Marshal(carAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(carAsset.ID, carAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put cars to world state. %v", err)
		}

		indexName := "color~owner~ID"
		colorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, carAsset.OwnerID, carAsset.ID})
		if err != nil {
			return err
		}

		value := []byte{0x00}
		err = ctx.GetStub().PutState(colorOwnerIndexKey, value)
		if err != nil {
			return err
		}
	}

	for _, personAsset := range personAssets {
		personAssetJSON, err := json.Marshal(personAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put persons to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) ReadPersonAsset(ctx contractapi.TransactionContextInterface, id string) (*PersonAsset, error) {
	personAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read person from world state: %v", err)
	}
	if personAssetJSON == nil {
		return nil, fmt.Errorf("the person asset %s does not exist", id)
	}

	var personAsset PersonAsset
	err = json.Unmarshal(personAssetJSON, &personAsset)
	if err != nil {
		return nil, err
	}

	return &personAsset, nil
}

func (s *SmartContract) ReadCarAsset(ctx contractapi.TransactionContextInterface, id string) (*CarAsset, error) {
	carAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read car from world state: %v", err)
	}
	if carAssetJSON == nil {
		return nil, fmt.Errorf("the car asset %s does not exist", id)
	}

	var carAsset CarAsset
	err = json.Unmarshal(carAssetJSON, &carAsset)
	if err != nil {
		return nil, err
	}

	return &carAsset, nil
}

func (s *SmartContract) GetCarsByColor(ctx contractapi.TransactionContextInterface, color string) ([]*CarAsset, error) {
	coloredCarIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color})
	if err != nil {
		return nil, err
	}

	defer coloredCarIter.Close()

	retList := make([]*CarAsset, 0)

	for i := 0; coloredCarIter.HasNext(); i++ {
		responseRange, err := coloredCarIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retCarID := compositeKeyParts[2]

		carAsset, err := s.ReadCarAsset(ctx, retCarID)
		if err != nil {
			return nil, err
		}

		retList = append(retList, carAsset)
	}

	return retList, nil
}

func (s *SmartContract) GetCarsByColorAndOwner(ctx contractapi.TransactionContextInterface, color string, ownerID string) ([]*CarAsset, error) {
	exists, err := s.PersonAssetExists(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("the person %v does not exist", ownerID)
	}

	coloredCarByOwnerIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color, ownerID})
	if err != nil {
		return nil, err
	}

	defer coloredCarByOwnerIter.Close()

	retList := make([]*CarAsset, 0)

	for i := 0; coloredCarByOwnerIter.HasNext(); i++ {
		responseRange, err := coloredCarByOwnerIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retCarID := compositeKeyParts[2]

		carAsset, err := s.ReadCarAsset(ctx, retCarID)
		if err != nil {
			return nil, err
		}

		retList = append(retList, carAsset)
	}

	return retList, nil
}

func (s *SmartContract) TransferCarAsset(ctx contractapi.TransactionContextInterface, id string, newOwnerID string, acceptMalfunction bool) (bool, error) {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return false, err
	}

	if carAsset.OwnerID == newOwnerID {
		return false, fmt.Errorf("Person %s is already the owner of the car!", newOwnerID)
	}

	buyer, err := s.ReadPersonAsset(ctx, newOwnerID)
	if err != nil {
		return false, err
	}

	seller, err := s.ReadPersonAsset(ctx, carAsset.OwnerID)
	if err != nil {
		return false, err
	}

	carPrice := float32(0)

	if carAsset.MalfunctionList == nil || len(carAsset.MalfunctionList) == 0 {
		carPrice = carAsset.Price
	} else if acceptMalfunction {
		malfuctionPrice := float32(0)
		for _, carMalfunction := range carAsset.MalfunctionList {
			malfuctionPrice += carMalfunction.RepairPrice
		}
		carPrice = carAsset.Price - malfuctionPrice
	} else {
		return false, fmt.Errorf("the buyer will not accept a malfunctioned car")
	}

	oldOwnerID := carAsset.OwnerID
	carAsset.OwnerID = newOwnerID

	if buyer.AmountOfMoneyOwned >= carPrice {
		buyer.AmountOfMoneyOwned -= carPrice
		seller.AmountOfMoneyOwned += carPrice
	} else {
		return false, fmt.Errorf("the buyer does not own enough money to purchase the car")
	}

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return false, err
	}

	buyerJSON, err := json.Marshal(buyer)
	if err != nil {
		return false, err
	}

	sellerJSON, err := json.Marshal(seller)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(buyer.ID, buyerJSON)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(seller.ID, sellerJSON)
	if err != nil {
		return false, err
	}

	indexName := "color~owner~ID"
	colorNewOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, newOwnerID, carAsset.ID})
	if err != nil {
		return false, err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(colorNewOwnerIndexKey, value)
	if err != nil {
		return false, err
	}

	colorOldOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, oldOwnerID, carAsset.ID})
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(colorOldOwnerIndexKey)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) AddCarMalfunction(ctx contractapi.TransactionContextInterface, id string, description string, repairPrice float32) error {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return err
	}

	newMalfunction := CarMalfunction{
		Description: description,
		RepairPrice: repairPrice,
	}

	carAsset.MalfunctionList = append(carAsset.MalfunctionList, newMalfunction)

	totalRepairPrice := float32(0)
	for _, carMalfunction := range carAsset.MalfunctionList {
		totalRepairPrice += carMalfunction.RepairPrice
	}

	if totalRepairPrice > carAsset.Price {
		return ctx.GetStub().DelState(id)
	}

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) ChangeCarColor(ctx contractapi.TransactionContextInterface, id string, newColor string) (string, error) {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldColor := carAsset.Color
	carAsset.Color = newColor

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return "", err
	}

	indexName := "color~owner~ID"
	newColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{newColor, carAsset.OwnerID, carAsset.ID})
	if err != nil {
		return "", err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(newColorOwnerIndexKey, value)
	if err != nil {
		return "", err
	}

	oldColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{oldColor, carAsset.OwnerID, carAsset.ID})
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().DelState(oldColorOwnerIndexKey)
	if err != nil {
		return "", err
	}

	return oldColor, nil
}

func (s *SmartContract) RepairCar(ctx contractapi.TransactionContextInterface, id string) error {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return err
	}

	personAsset, err := s.ReadPersonAsset(ctx, carAsset.OwnerID)
	if err != nil {
		return err
	}

	repairPriceSum := float32(0)
	for _, carMalfunction := range carAsset.MalfunctionList {
		repairPriceSum += carMalfunction.RepairPrice
		if repairPriceSum > personAsset.AmountOfMoneyOwned {
			return fmt.Errorf("The owner of the car cannot afford to pay the car repair price")
		}
	}

	carAsset.MalfunctionList = []CarMalfunction{}
	personAsset.AmountOfMoneyOwned -= repairPriceSum

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return err
	}

	personAssetJSON, err := json.Marshal(personAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) PersonAssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	personAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read person asset from world state: %v", err)
	}

	return personAssetJSON != nil, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating cars-and-persons chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting cars-and-persons chaincode: %v", err)
	}
}
