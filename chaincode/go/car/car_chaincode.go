
package main

import (
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"crypto/x509"
	"strings"
	"encoding/pem"
)

type CarChaincode struct {
}

type Part struct {
	Maker  	    	string 				`json:"maker"`
	Id  	    	string 				`json:"id"`
}

type Car struct {
	Vin 			string 				`json:"vin"`
	Engine 			Part 				`json:"engine"`
	Body 			Part 				`json:"body"`
}

var logger = shim.NewLogger("CarChaincode")

const indexName = `Car`

func (t *CarChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *CarChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	if function == "order" {
		return t.order(stub, args)
	} else if function == "supply" {
		return t.supply(stub, args)
	} else if function == "setMaker" {
		return t.setMaker(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	}

	return pb.Response{Message:"unknown function",Status:400}
}

func (t *CarChaincode) order(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	vin := args[0]

	value, err := json.Marshal(Car{Vin:vin})
	if err != nil {
		return shim.Error("cannot marshal")
	}

	key, err := stub.CreateCompositeKey(indexName, []string{vin})
	if err != nil {
		return shim.Error("cannot create composite key")
	}

	err = stub.PutState(key, []byte(value))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *CarChaincode) setMaker(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	cert, _ := stub.GetCreator()
	org := getOrganization(cert)

	if org != "factory" {
		return pb.Response{Message:"only factory authorized to call setMaker", Status:401}
	}

	vin := args[0]
	part := args[1]
	maker := args[2]

	car, err := getCar(stub, vin)
	if err != nil {
		return shim.Error("cannot get car")
	}

	if part == "engine" {
		car.Engine.Maker = maker
	} else {
		return pb.Response{Message:"cannot set anything but engine", Status:400}
	}

	err = putCar(stub, car)
	if err != nil {
		return shim.Error("cannot put car")
	}

	return shim.Success(nil)
}

func (t *CarChaincode) supply(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	cert, _ := stub.GetCreator()
	org := getOrganization(cert)

	if org != "rr" && org != "ferrari" && org != "gm" {
		return pb.Response{Message:"only supplier authorized to call supply", Status:401}
	}

	vin := args[0]
	part := args[1]
	id := args[2]

	car, err := getCar(stub, vin)
	if err != nil {
		return shim.Error("cannot get car")
	}

	if org == "gm" && part == "body" {
		car.Body.Maker = org
		car.Body.Id = id
	} else if car.Engine.Maker == org && part == "engine" {
		car.Engine.Id = id
	} else {
		return pb.Response{Message:"wrong org or part", Status:401}
	}

	err = putCar(stub, car)
	if err != nil {
		return shim.Error("cannot put car")
	}

	return shim.Success(nil)
}

func (t *CarChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	it, err := stub.GetStateByPartialCompositeKey(indexName, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	defer it.Close()

	arr := []Car{}
	for it.HasNext() {
		next, err := it.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var car Car
		err = json.Unmarshal(next.Value, &car)
		if err != nil {
			return shim.Error(err.Error())
		}

		arr = append(arr, car)
	}

	ret, err := json.Marshal(arr)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(ret)
}

func getCar(stub shim.ChaincodeStubInterface, vin string) (Car, error) {
	key, err := stub.CreateCompositeKey(indexName, []string{vin})
	if err != nil {
		return Car{}, err
	}

	value, err := stub.GetState(key)
	if err != nil {
		return Car{}, err
	}

	var car Car
	err = json.Unmarshal(value, &car)
	if err != nil {
		return Car{}, err
	}

	return car, nil
}

func putCar(stub shim.ChaincodeStubInterface, car Car) error {
	key, err := stub.CreateCompositeKey(indexName, []string{car.Vin})
	if err != nil {
		return err
	}

	value, err := json.Marshal(car)
	if err != nil {
		return err
	}

	err = stub.PutState(key, []byte(value))
	if err != nil {
		return err
	}

	return nil
}

func getOrganization(certificate []byte) string {
	data := certificate[strings.Index(string(certificate), "-----") : strings.LastIndex(string(certificate), "-----")+5]
	block, _ := pem.Decode([]byte(data))
	cert, _ := x509.ParseCertificate(block.Bytes)
	organization := cert.Issuer.Organization[0]
	logger.Info("getOrganization: " + organization)

	ret := strings.Split(organization, ".")[0]

	return ret
}

func main() {
	err := shim.Start(new(CarChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
