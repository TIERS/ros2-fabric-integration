package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a asset
type ObjectDetection struct {
	Object      string `json:"Object"`
	ClassID     string `json:"ClassID"`
	Probability string `json:"Probability"`
	XPosition   string `json:"XPosition"`
	YPosition   string `json:"YPosition"`
	ZPosition   string `json:"ZPosition"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	objectdetections := []ObjectDetection{}

	for _, objectdetection := range objectdetections {
		objectdetectionJSON, err := json.Marshal(objectdetection)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(objectdetection.Object, objectdetectionJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateObjectDetection issues a new asset to the world state with given details.
func (s *SmartContract) CreateObjectDetection(ctx contractapi.TransactionContextInterface, object string, classid string, probability string, xposition string, yposition string, zposition string) error {
	exists, err := s.ObjectDetectionExists(ctx, object)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the objectdetection %s already exists", object)
	}

	objectdetection := ObjectDetection{
		Object:      object,
		ClassID:     classid,
		Probability: probability,
		XPosition:   xposition,
		YPosition:   yposition,
		ZPosition:   zposition,
	}
	objectdetectionJSON, err := json.Marshal(objectdetection)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(object, objectdetectionJSON)
}

// ReadObjectDetection returns the asset stored in the world state with given details.
func (s *SmartContract) ReadObjectDetection(ctx contractapi.TransactionContextInterface, object string) (*ObjectDetection, error) {
	objectdetectionJSON, err := ctx.GetStub().GetState(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if objectdetectionJSON == nil {
		return nil, fmt.Errorf("the objectdetection %s does not exist", object)
	}

	var objectdetection ObjectDetection
	err = json.Unmarshal(objectdetectionJSON, &objectdetection)
	if err != nil {
		return nil, err
	}

	return &objectdetection, nil
}

// UpdateObjectDetection updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateObjectDetection(ctx contractapi.TransactionContextInterface, object string, classid string, probability string, xposition string, yposition string, zposition string) error {
	exists, err := s.ObjectDetectionExists(ctx, object)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the objectdetection %s does not exist", object)
	}

	// overwriting original asset with new asset
	objectdetection := ObjectDetection{
		Object:      object,
		ClassID:     classid,
		Probability: probability,
		XPosition:   xposition,
		YPosition:   yposition,
		ZPosition:   zposition,
	}
	objectdetectionJSON, err := json.Marshal(objectdetection)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(object, objectdetectionJSON)
}

// DeleteObjectDetection deletes an given asset from the world state.
func (s *SmartContract) DeleteObjectDetection(ctx contractapi.TransactionContextInterface, object string) error {
	exists, err := s.ObjectDetectionExists(ctx, object)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", object)
	}

	return ctx.GetStub().DelState(object)
}

// ObjectDetectionExists returns true when asset with given ID exists in world state
func (s *SmartContract) ObjectDetectionExists(ctx contractapi.TransactionContextInterface, object string) (bool, error) {
	objectdetectionJSON, err := ctx.GetStub().GetState(object)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return objectdetectionJSON != nil, nil
}

// GetAllObjectDetections returns all assets found in world state
func (s *SmartContract) GetAllObjectDetections(ctx contractapi.TransactionContextInterface) ([]*ObjectDetection, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var objectdetections []*ObjectDetection
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var objectdetection ObjectDetection
		err = json.Unmarshal(queryResponse.Value, &objectdetection)
		if err != nil {
			return nil, err
		}
		objectdetections = append(objectdetections, &objectdetection)
	}

	return objectdetections, nil
}
