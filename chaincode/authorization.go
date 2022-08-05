package main

import (
  "encoding/json"
  "fmt"
  "log"
  "time"

  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
   type SmartContract struct {
      contractapi.Contract
    }

// Asset describes basic details of what makes up a simple asset
    type Asset struct {
      ID             string `json:"ID"`
      Role           string `json:"role"`
      TimeStamp      string `json:"timestamp"`
      PublicKey      string `json:"publickey"`
    }


// CreateAsset issues a new asset to the world state with given details.
   func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, id string, role string, publickey string) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
      return err
    }
    if exists {
      return fmt.Errorf("the asset %s already exists", id)
    }
    if (role != "controller" && role != "processor"){
      return fmt.Errorf("wrong role %s, try controller or processor", role)
    }

    t := time.Now()
    timestamp := t.Format("2006-01-02 15:04:05")

    asset := Asset{
      ID:             id,
      Role:           role,
      TimeStamp:      timestamp,
      PublicKey:      publickey,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
      return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
  }

// ReadAsset returns the asset stored in the world state with given id.
   func (s *SmartContract) Read(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
    assetJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
      return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if assetJSON == nil {
      return nil, fmt.Errorf("the asset %s does not exist", id)
    }

    var asset Asset
    err = json.Unmarshal(assetJSON, &asset)
    if err != nil {
      return nil, err
    }

    return &asset, nil
  }


  // DeleteAsset deletes an given asset from the world state.
  func (s *SmartContract) Delete(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
      return err
    }
    if !exists {
      return fmt.Errorf("the asset %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
  }

// AssetExists returns true when asset with given ID exists in world state
   func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
    assetJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
      return false, fmt.Errorf("failed to read from world state: %v", err)
    }

    return assetJSON != nil, nil
  }




// GetAllAssets returns all assets found in world state
   func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
// range query with empty string for startKey and endKey does an
// open-ended query of all assets in the chaincode namespace.
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
      return nil, err
    }
    defer resultsIterator.Close()

    var assets []*Asset
    for resultsIterator.HasNext() {
      queryResponse, err := resultsIterator.Next()
      if err != nil {
        return nil, err
      }

      var asset Asset
      err = json.Unmarshal(queryResponse.Value, &asset)
      if err != nil {
        return nil, err
      }
      assets = append(assets, &asset)
    }

    return assets, nil
  }

  //Change the public key parameter of the admin in the ledger
     func (s *SmartContract) NewPublicKey(ctx contractapi.TransactionContextInterface, id string, public_key string) error {
      asset, err := s.Read(ctx, id)
      if err != nil {
        return err
      }

      asset.PublicKey = public_key
      assetJSON, err := json.Marshal(asset)
      if err != nil {
        return err
      }

      return ctx.GetStub().PutState(id, assetJSON)
    }

    // Returns the role of an admin (controller or processor)
       func (s *SmartContract) GetRole(ctx contractapi.TransactionContextInterface, id string) (string, error) {
        asset, err := s.Read(ctx, id)
        if err != nil {
          return "", err
        }

        var role string
        role = asset.Role

        return role, nil
      }


  func main() {
    assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
      log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
    }

    if err := assetChaincode.Start(); err != nil {
      log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
    }
  }
