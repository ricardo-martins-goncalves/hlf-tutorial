package main

import (
  "encoding/json"
  "fmt"
  "log"
  "time"
  "strings"

  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
   type SmartContract struct {
      contractapi.Contract
    }

// Asset describes basic details of what makes up a simple asset
   type Asset struct {
      ID             string `json:"ID"`
      Consents       string `json:"consents"`
      TimeStamp      string `json:"timestamp"`
      CID            string `json:"cid"`
      Hash           string `json:"hash"`
      Surveys        string `json:"surveys"`
      Deleted        bool   `json:"deleted"`
    }


// CreateAsset issues a new asset to the world state with given details.
   func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, consents string, cid string, hash string) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
      return err
    }
    if exists {
      return fmt.Errorf("the asset %s already exists", id)
    }
    t := time.Now()
    timestamp := t.Format("2006-01-02 15:04:05")

    asset := Asset{
      ID:             id,
      Consents:       consents,
      TimeStamp:      timestamp,
      CID:            cid,
      Surveys:        "",
      Hash:           hash,
      Deleted:        false,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
      return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
  }

// ReadAsset returns the asset stored in the world state with given id.
   func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
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

// UpdateAsset updates an existing asset in the world state with provided parameters.
   func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, consents string, cid string, hash string) error {
    exists, err := s.AssetExists(ctx, id)
    if err != nil {
      return err
    }
    if !exists {
      return fmt.Errorf("the asset %s does not exist", id)
    }
    t := time.Now()
    timestamp := t.Format("2006-01-02 15:04:05")

    // overwriting original asset with new asset
    asset := Asset{
      ID:             id,
      Consents:       consents,
      TimeStamp:      timestamp,
      CID:            cid,
      Hash:           hash,
      Deleted:         false,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
      return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
  }

  // DeleteAsset deletes an given asset from the world state.
  func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
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

//When the correspondent ipfs file is deleted, update the deleted parameter.
   func (s *SmartContract) SetDeleted(ctx contractapi.TransactionContextInterface, id string) error {
    asset, err := s.ReadAsset(ctx, id)
    if err != nil {
      return err
    }

    asset.Deleted = true
    assetJSON, err := json.Marshal(asset)
    if err != nil {
      return err
    }

    return ctx.GetStub().PutState(id, assetJSON)
  }

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

  //Add cid of a ipfs file to the list of cids in the survey
      func (s *SmartContract) AddCID(ctx contractapi.TransactionContextInterface, asset_id string, survey_id string, cid string) error {
       asset, err := s.ReadAsset(ctx, asset_id)
       if err != nil {
         return err
       }

       //search the surveys in the list
       surveys := asset.Surveys

       index := strings.Index(cid, ";") + strings.Index(cid, "_")
       if index != -2 {
         return fmt.Errorf("Invalid survey, CID pair: %s, %s ", survey_id, cid)
       }

       index = strings.Index(surveys, survey_id + "_" + cid)
       if index != -1 {
         return fmt.Errorf("The pair survey, CID %s %s is already on the list of Surveys", survey_id, cid)
       }


       asset.Surveys = surveys + survey_id + "_" + cid + ";"

       assetJSON, err := json.Marshal(asset)
       if err != nil {
         return err
       }

       return ctx.GetStub().PutState(asset_id, assetJSON)
     }

     //Delete cid of the list of cids of a survey
        func (s *SmartContract) RemoveCID(ctx contractapi.TransactionContextInterface, asset_id string, survey_id string, cid string) error {
         asset, err := s.ReadAsset(ctx, asset_id)
         if err != nil {
           return err
         }

         //search the cid in the list and deleted if found
         surveys := asset.Surveys
         index := strings.Index(surveys, survey_id + "_" +cid)
         if index == -1 {
           return fmt.Errorf("The pair Survey, CID %s, %s does not exist in this survey", survey_id, cid)
         }

         surveys = strings.ReplaceAll(surveys, survey_id + "_" + cid + ";", "")
         asset.Surveys = surveys

         assetJSON, err := json.Marshal(asset)
         if err != nil {
           return err
         }

         return ctx.GetStub().PutState(asset_id, assetJSON)
       }

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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




  func main() {
    assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
      log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
    }

    if err := assetChaincode.Start(); err != nil {
      log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
    }
  }
