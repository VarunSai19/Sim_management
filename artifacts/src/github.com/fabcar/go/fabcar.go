package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"encoding/base64"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

type SmartContract struct {
	contractapi.Contract
}

var logger = flogging.MustGetLogger("fabcar_cc")

type CallDetails struct{
	callBegin time.Time `json:"callBegin"`
	callEnd   time.Time `json:"callEnd"`
	callCharges  float32 `json:"callCharges"`
}

type CSPData struct{
	name string `json:"name"`
	region   string `json:"region"`
	overageRate float32 `json:"overageRate"`
	roamingRate  float32 `json:"roamingRate"`
	latitude string `json:"latitude"`
	longitude string `json:"longitude"`
	Doc_type string `json:"Doc_type"`
}

type SimData struct{
	publicKey  string `json:"publicKey"`
	msisdn string `json:"msisdn"`
	address   string `json:"address"`
	homeOperatorName string `json:"homeOperatorName"`
	roamingPartnerName   string `json:"roamingPartnerName"`
	isRoaming   string `json:"isRoaming"`
	location   string `json:"location"`
	longitude   string `json:"longitude"`
	latitude   string `json:"latitude"`
	roamingRate  float32 `json:"roamingRate"`
	overageRate  float32 `json:"overageRate"`	
	callDetails   []CallDetails `json:"callDetails"`
	isValid   string `json:"isValid"`
	overageThreshold   float32 `json:"overageThreshold"`
	overageFlag   string `json:"overageFlag"`
	allowOverage string `json:"allowOverage"`
	Doc_type string `json:"Doc_type"`
}

type AadharData struct {
	AadharNumber   string `json:"AadharNumber"`
	Address    string `json:"Address"`
	DateOfBirth   string `json:"DateOfBirth"`
	Name   string `json:"Name"`
	Gender   string `json:"Gender"`
}

type DrivingLicence struct {
	LicenceNumber  string `json:"LicenceNumber"`
	Address    string `json:"Address"`
	DateOfBirth   string `json:"DateOfBirth"`
	Name   string `json:"Name"`
	Gender   string `json:"Gender"`
	LicenceValidity   string `json:"LicenceValidity"`
}

type Car struct {
	ID      string `json:"id"`
	Make    string `json:"make"`
	Model   string `json:"model"`
	Color   string `json:"color"`
	Owner   string `json:"owner"`
	AddedAt uint64 `json:"addedAt"`	
}

func (s *SmartContract) assetExist(ctx contractapi.TransactionContextInterface, Id string) bool {
	if len(ID) == 0 {
		return false
	}
	dataAsBytes, err := ctx.GetStub().GetState(ID)

	if err != nil {
		return false
	}

	if dataAsBytes == nil {
		return false
	}

	return true
}

func (s *SmartContract) ReadCSPData(ctx contractapi.TransactionContextInterface, ID string) (*CSPData, error) {
	if len(ID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract Id")
	}
	dataAsBytes, err := ctx.GetStub().GetState(ID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if dataAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ID)
	}
	data := new(CSPData)
	_ = json.Unmarshal(dataAsBytes, data)

	return data, nil
}

func (s *SmartContract) ReadSimData(ctx contractapi.TransactionContextInterface, ID string) (*SimData, error) {
	if len(ID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract Id")
	}
	dataAsBytes, err := ctx.GetStub().GetState(ID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if dataAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ID)
	}
	data := new(SimData)
	_ = json.Unmarshal(dataAsBytes, data)

	return data, nil
}

func (s *SmartContract) checkForFraud(ctx contractapi.TransactionContextInterface, simpublickey string) (bool,error) {
	exist := s.assetExist(ctx,simpublickey)
	if !exist {
		return false,fmt.Errorf("Sim doesnt exist")
	}
	data := s.ReadSimData(simpublickey)
	if data.isValid == "fraud" {
		return true,nil
	}
	return false,nil
}

func (s *SmartContract) CreateCSP(ctx contractapi.TransactionContextInterface, Data string) (string, error) {
	if len(Data) == 0 {
		return "", fmt.Errorf("Please pass the correct data")
	}

	var data CSPData
	err := json.Unmarshal([]byte(Data), &data)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling Data. %s", err.Error())
	}

	exist := s.assetExist(ctx,data.name)

	if exist {
		return "",fmt.Errorf("CSP data already exist.")
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(data.name, dataAsBytes)
}


func (s *SmartContract) CreateSubscriberSim(ctx contractapi.TransactionContextInterface, Data string) (string, error) {
	if len(Data) == 0 {
		return "", fmt.Errorf("Please pass the correct data")
	}

	var data SimData
	err := json.Unmarshal([]byte(Data), &data)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling Data. %s", err.Error())
	}

	exist := s.assetExist(ctx,data.publicKey)
	if exist {
		return "",fmt.Errorf("public key is already exist.")
	}

	exist = s.assetExist(ctx,data.homeOperatorName)
	if !exist {
		return "",fmt.Errorf("Home operator doesnt exist.")
	}

	csp_data,err := s.ReadCSPData(ctx,data.homeOperatorName)

	data.location = csp_data.region
	data.latitude = csp_data.latitude
	data.longitude = csp_data.longitude

	if data.RoamingPartnerName {
		exist = s.assetExist(ctx,data.RoamingPartnerName)
		if !exist {
			return "",fmt.Errorf("Roaming Partner doesnt exist.")
		}
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(data.publicKey, dataAsBytes)
}

func (s *SmartContract) UpdateCSP(ctx contractapi.TransactionContextInterface, Data string) error {
	if len(Data) == 0 {
		return fmt.Errorf("Please pass the correct data")
	}

	var newdata CSPData
	err := json.Unmarshal([]byte(Data), &newdata)
	if err != nil {
		return fmt.Errorf("Failed while unmarshling Data. %s", err.Error())
	}

	exist := s.assetExist(ctx,newData.name)
	if !exist {
		fmt.Errorf("CSP data does not exist.")
	}

	dataAsBytes, err := json.Marshal(newdata)
	if err != nil {
		return fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().PutState(newdata.name, dataAsBytes)
}

func (s *SmartContract) DeleteCSP(ctx contractapi.TransactionContextInterface, Id string) error {
	if len(Id) == 0 {
		return fmt.Errorf("Please pass the correct data")
	}

	exist := s.assetExist(ctx,Id)
	if !exist {
		return fmt.Errorf("CSP doesnt exist.")
	}

	data := s.ReadSimData(ctx,Id)
	if data.Doc_type != "CSP"{
		return fmt.Errorf("CSP doesnt exist.")
	}

	AllCSP_simData,err = s.findAllSubscriberSimsForCSP(ctx,Id)

	if len(AllCSP_simData) > 0 {
		fmt.Errorf("The CSP can not be deleted as the following sims are currently in its network")
	}

	return ctx.GetStub().DelState(data.publicKey)
}

func (s *SmartContract) UpdateSubscriberSim(ctx contractapi.TransactionContextInterface, Data string) error {
	if len(Data) == 0 {
		return fmt.Errorf("Please pass the correct data")
	}

	var data SimData
	err := json.Unmarshal([]byte(Data), &data)
	if err != nil {
		return fmt.Errorf("Failed while unmarshling Data. %s", err.Error())
	}

	exist := s.assetExist(ctx,data.publicKey)
	if !exist {
		return fmt.Errorf("public key doesnt exist.")
	}

	exist = s.assetExist(ctx,data.homeOperatorName)
	if !exist {
		return fmt.Errorf("Home operator doesnt exist.")
	}

	if data.RoamingPartnerName {
		exist = s.assetExist(ctx,data.RoamingPartnerName)
		if !exist {
			return fmt.Errorf("Roaming Partner doesnt exist.")
		}
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().PutState(data.publicKey, dataAsBytes)
}


func (s *SmartContract) DeleteSubscriberSim(ctx contractapi.TransactionContextInterface, Id string) error {
	if len(Id) == 0 {
		return fmt.Errorf("Please pass the correct data")
	}

	exist := s.assetExist(ctx,Id)
	if !exist {
		return fmt.Errorf("public key doesnt exist.")
	}

	data := s.ReadSimData(ctx,Id)
	if data.Doc_type != "SubscriberSim"{
		return fmt.Errorf("public key doesnt exist.")
	}
	return ctx.GetStub().DelState(data.publicKey)
}

func (s *SmartContract) MoveSim(ctx contractapi.TransactionContextInterface, publicKey string,location string) error {
	if len(publicKey) == 0 {
		return "", fmt.Errorf("Please pass the correct data")
	}

	var data SimData
	err := json.Unmarshal([]byte(Data), &data)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling Data. %s", err.Error())
	}

	exist := s.assetExist(ctx,data.publicKey)
	if !exist {
		return "",fmt.Errorf("public key doesnt exist.")
	}

	data.location = location

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().PutState(data.publicKey, dataAsBytes)
}

func (s *SmartContract) UpdateRate(ctx contractapi.TransactionContextInterface, publicKey string, RoamingPartnerName string) error {
	if len(publicKey) == 0 {
		return fmt.Errorf("Please pass the correct data")
	}

	exist := s.assetExist(ctx,publicKey)
	if !exist {
		fmt.Errorf("Sim does not exist.")
	}

	data,err := s.ReadSimData(publicKey)

	if err != nil {
		fmt.Errorf("Error while reading the asset.")
	}

	if data.isValid == "fraud" {
		fmt.Errorf("The user public key is marked as as fraudulent because the msisdn specified by this user is already in use. No calls can be made by this user.");
	}

	if data.homeOperatorName == RoamingPartnerName && data.isRoaming == true {
		data.roamingPartnerName = ''
		data.isRoaming = false
		data.roamingRate = 0
		data.overageRate = 0
	}
	else if data.homeOperatorName != RoamingPartnerName{
		exist = s.assetExist(ctx,RoamingPartnerName)
		if !exist {
			fmt.Errorf("Roaming partner does not exist.")
		}

		roamingData,err = s.ReadCSPData(ctx,RoamingPartnerName)
		data.roamingPartnerName = roamingData.name
		data.isRoaming = true
		data.roamingRate = roamingData.roamingRate
		data.overageRate = roamingData.overageRate
	}

	dataAsBytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}

	return ctx.GetStub().PutState(data.publicKey, dataAsBytes)
}

func (s *SmartContract) discovery(ctx contractapi.TransactionContextInterface, publicKey string) (string,error) {
	exist := s.assetExist(ctx,data.publicKey)
	if !exist {
		return fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)

	exist = s.assetExist(ctx,data.homeOperatorName)

	if !exist {
		return fmt.Errorf("Home operator doesnt exist.")
	}
	var operator
	Homedata,err := s.ReadCSPData(ctx,data.homeOperatorName)
	if Homedata.region != simdata.location {
		queryString := fmt.Sprintf(`{"selector":{"Doc_type":"CSP","region":"%s"}}`,simdata.location)
		operators,err := s.getQueryResultSimData(ctc,queryString)
		if len(operators) == 0 {
			fmt.Errorf("No operators found for the location.")
		}
		else {
			operator = operators[0]
		}
		
	}
	else{
		operator = data.homeOperatorName
	}
	return operator,nil
}

func (s *SmartContract) authentication(ctx contractapi.TransactionContextInterface, publicKey string) error {
	exist := s.assetExist(ctx,data.publicKey)
	if !exist {
		return fmt.Errorf("public key does not exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	// Checking for Sim Cloning and we are assigning it as Fraud. 
	queryString := fmt.Sprintf(`{"selector":{"Doc_type":"SubscriberSim","isValid":"active","publicKey":{"$nin":["%s"]},"msisdn":"%s"}}`,publicKey,simdata.msisdn)
	queryRes,err := s.getQueryResultSimData(ctc,queryString)
	var valid
	if len(queryRes) > 0 {
		valid = "fraud"
	}
	else{
		valid = "active"
	}
	if simdata.isValid != valid{
		simdata.isValid = valid
		dataAsBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("Failed while marshling Data. %s", err.Error())
		}
		ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
	}
	return nil
}


func (s *SmartContract) VerifyUser(ctx contractapi.TransactionContextInterface, publicKey string) (string,string,error) {
	is_fraud,err := s.checkForFraud(ctx,publicKey)

	if err != nil {
		return "","",fmt.Errorf("The sim doesnt exist with this public key.")
	}

	if(is_fraud) {
		return "","",fmt.Errorf("This is a fraud sim.")
	}

	return s.checkForOverage(ctc,publicKey)
}

func (s *SmartContract) checkForOverage(ctx contractapi.TransactionContextInterface, publicKey string) (string,string,error) {
	exist := s.assetExist(ctx,publicKey)
	if !exist {
		return "","",fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	if simdata.overageFlag == "true" {
		return asset.overageFlag, asset.allowOverage,nil;
	}

	var calldetails = simdata.CallDetails
	var total_charge = 0

	for _,calldetail := range calldetails {
		total_charge += calldetail.callCharges
	}
	if total_charge + simdata.roamingRate > simdata.overageThreshold {
		simdata.overageFlag = "true"
		dataAsBytes, err := json.Marshal(simdata)
		if err != nil {
			return fmt.Errorf("Failed while marshling Data. %s", err.Error())
		}
		return "true",simdata.allowOverage,ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
	}
	else{
		return simdata.overageFlag, simdata.allowOverage,nil
	}
	
}


func (s *SmartContract) setOverageFlag(ctx contractapi.TransactionContextInterface, publicKey string, allowOverage string) error {
	exist := s.assetExist(ctx,publicKey)
	if !exist {
		return fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	if simdata.overageFlag == "true" {
		return nil;
	}

	if simdata.overageFlag == "true" && simdata.allowOverage == '' {
		simdata.allowOverage = allowOverage
		dataAsBytes, err := json.Marshal(simdata)
		if err != nil {
			return fmt.Errorf("Failed while marshling Data. %s", err.Error())
		}
		return ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
	}	
}

func (s *SmartContract) callOut(ctx contractapi.TransactionContextInterface, publicKey string) error {
	exist := s.assetExist(ctx,publicKey)
	if !exist {
		return fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	if simdata.overageFlag == "true" {
		return nil;
	}

	if simdata.overageFlag == "true" && simdata.allowOverage == "false" {
		return fmt.Errorf("No further calls will be allowed as the user has reached the overage threshold and has denied the overage charges.")
	}	

	now := time.Now()
	var calldetail = new(CallDetails)
	calldetail.callBegin = now
	calldetail.callEnd = now.Add(time.Duration(-1) * time.Minute)
	calldetail.callCharges = 0
	list = append(list,*calldetail)
	// time.Unix(time.Now().Unix(),0)

	dataAsBytes, err := json.Marshal(simdata)
	if err != nil {
		return fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}
	return ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
}

func (s *SmartContract) callEnd(ctx contractapi.TransactionContextInterface, publicKey string) error {
	exist := s.assetExist(ctx,publicKey)
	if !exist {
		return fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	
	if simdata.isValid == "fraud" {
		return fmt.Errorf("This user has been marked as fraudulent because the msisdn specified by this user is already in use. No calls can be made by this user.")
	}

	last_index := len(simdata.CallDetails)-1
	calldetail := simdata.CallDetails[last_index]
	begin := calldetail.callBegin
	end := calldetail.callEnd

	if begin.Before(end) {
		fmt.Errorf("No ongoing call for the user was found. Can not continue with callEnd process.")
	}
	// time.Unix(time.Now().Unix(),0)
	simdata.CallDetails[last_index].callEnd = time.Now()
	dataAsBytes, err := json.Marshal(simdata)
	if err != nil {
		return fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}
	return ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
}

func (s *SmartContract) callPay(ctx contractapi.TransactionContextInterface, publicKey string) error {
	exist := s.assetExist(ctx,publicKey)
	if !exist {
		return fmt.Errorf("public key does not already exist.")
	}

	simdata,err = s.ReadSimData(ctx,publicKey)
	var rate
	if simdata.overageFlag === "true" {
		rate = simdata.overageRate
	}
	else{
		rate = simdata.roamingRate
	}

	last_index := len(simdata.CallDetails)-1
	calldetail := simdata.CallDetails[last_index]
	begin := calldetail.callBegin
	end := calldetail.callEnd

	duration := end.Sub(begin).Minutes()
	simdata.CallDetails[last_index].callCharges = duration*rate
	dataAsBytes, err := json.Marshal(simdata)
	if err != nil {
		return fmt.Errorf("Failed while marshling Data. %s", err.Error())
	}
	return ctx.GetStub().PutState(simdata.publicKey, dataAsBytes)
}


func (s *SmartContract) GetHistoryForAsset(ctx contractapi.TransactionContextInterface, ID string) (string, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(ID)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return string(buffer.Bytes()), nil
}

func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	// x509::CN=telco-admin,OU=o 
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	res := string(decodeID)
	i:=0
	id:=""
	for ;i<len(res);i++{
		if res[i] == '='{
			break	
		}
	}
	for i=i+1;i<len(res);i++{
		if res[i] == ','{
			break	
		} 
		id += string(res[i])
	} 
	return id, nil
}


func (s *SmartContract) getQueryResultData(ctx contractapi.TransactionContextInterface, queryString string) ([]CSPData, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	results := []CSPData{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		newData := new(CSPData)
		
		fmt.Print("Responce is ",response.Value,"\n")
		err = json.Unmarshal(response.Value, newData)
		if err == nil {
			results = append(results, *newData)
		}
	}
	return results, nil
}

func (s *SmartContract) QueryAllCSP(ctx contractapi.TransactionContextInterface, queryString string) ([]CSPData, error) {
	err := ctx.GetClientIdentity().AssertAttributeValue("usertype", "CSP")
	if err != nil {
		return nil,fmt.Errorf("submitting client not authorized to perform this task.")
	}
	return s.getQueryResultData(ctx,queryString)
}

func (s *SmartContract) getQueryResultSimData(ctx contractapi.TransactionContextInterface, queryString string) ([]SimData, error) {

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	
	results := []SimData{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		newData := new(SimData)
		
		fmt.Print("Responce is ",response.Value,"\n")
		err = json.Unmarshal(response.Value, newData)
		if err == nil {
			results = append(results, *newData)
		}
	}
	return results, nil
}

func (s *SmartContract) QueryAllSimData(ctx contractapi.TransactionContextInterface, queryString string) ([]SimData, error) {
	err := ctx.GetClientIdentity().AssertAttributeValue("usertype", "CSP")
	if err != nil {
		return nil,fmt.Errorf("submitting client not authorized to perform this task.")
	}

	return s.getQueryResultSimData(ctx,queryString)
}

func (s *SmartContract) findAllSubscriberSimsForCSP(ctx contractapi.TransactionContextInterface, csp_name string) ([]SimData, error) {
	err := ctx.GetClientIdentity().AssertAttributeValue("usertype", "CSP")
	if err != nil {
		return nil,fmt.Errorf("submitting client not authorized to perform this task.")
	}
	// var csp_name = "Airtel"
	queryString := fmt.Sprintf(`{"selector":{"Doc_type":"SubscriberSim","$or":[{"homeOperatorName":"%s"},{"roamingPartnerName":"%s"}]}}`,csp_name,csp_name)
	return s.getQueryResultSimData(ctx,queryString)
}


func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, carData string) (string, error) {

	if len(carData) == 0 {
		return "", fmt.Errorf("Please pass the correct car data")
	}

	var car Car
	err := json.Unmarshal([]byte(carData), &car)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling car. %s", err.Error())
	}

	carAsBytes, err := json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling car. %s", err.Error())
	}

	ctx.GetStub().SetEvent("CreateAsset", carAsBytes)

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(car.ID, carAsBytes)
}


//
func (s *SmartContract) UpdateCarOwner(ctx contractapi.TransactionContextInterface, carID string, newOwner string) (string, error) {

	if len(carID) == 0 {
		return "", fmt.Errorf("Please pass the correct car id")
	}

	carAsBytes, err := ctx.GetStub().GetState(carID)

	if err != nil {
		return "", fmt.Errorf("Failed to get car data. %s", err.Error())
	}

	if carAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", carID)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	car.Owner = newOwner

	carAsBytes, err = json.Marshal(car)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling car. %s", err.Error())
	}

	//  txId := ctx.GetStub().GetTxID()

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(car.ID, carAsBytes)

}


func (s *SmartContract) GetCarById(ctx contractapi.TransactionContextInterface, carID string) (*Car, error) {
	if len(carID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract Id")
		// return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carAsBytes, err := ctx.GetStub().GetState(carID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if carAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", carID)
	}

	car := new(Car)
	_ = json.Unmarshal(carAsBytes, car)

	return car, nil

}

func (s *SmartContract) DeleteCarById(ctx contractapi.TransactionContextInterface, carID string) (string, error) {
	if len(carID) == 0 {
		return "", fmt.Errorf("Please provide correct contract Id")
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().DelState(carID)
}



func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}

}
