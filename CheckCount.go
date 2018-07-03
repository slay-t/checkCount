package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"os"
)

type CheckCountEvent struct {
	Details struct {
		Parameters struct {
			CustomerNumber string `json:"CustomerNumber"`
		} `json:"Parameters"`
	} `json:"Details"`
}

type CheckCountResponse struct {
	CustomerCount int    `json:"CustomerCount"`
	Message       string `dynamo:"message"`
}

type merryCount struct {
	CustomerNumber string `dynamo:"customer_number"`
	CallCount      int    `dynamo:"call_count"`
}

func CheckCount(event CheckCountEvent) (CheckCountResponse, error) {
	session, err := session.NewSession()
	if err != nil {
		return CheckCountResponse{Message: "session create failed"}, err
	}
	db := dynamo.New(session, aws.NewConfig().WithRegion(os.Getenv("Region")))
	merryCountTable := db.Table("merry_count")

	var merryCountData merryCount
	if err := merryCountTable.Get("customer_number", event.Details.Parameters.CustomerNumber).One(&merryCountData); err != nil {
		merryCountData.CustomerNumber = event.Details.Parameters.CustomerNumber
	}

	merryCountData.CallCount++
	responseData := CheckCountResponse{CustomerCount: merryCountData.CallCount,
		Message: fmt.Sprintf("customer number:%s count:%d", merryCountData.CustomerNumber, merryCountData.CallCount)}

	if merryCountData.CallCount == 3 {
		merryCountData.CallCount = 0
	}

	if err := merryCountTable.Put(merryCountData).Run(); err != nil {
		return CheckCountResponse{Message: "table put failed"}, errors.New("merry_count put data failed")
	}

	return responseData, nil
}

func main() {
	lambda.Start(CheckCount)
}
