package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"log"
	"os"
	"strconv"
)

type CheckCountEvent struct {
	Details struct {
		Parameters struct {
			CheckCount string `json:"checkCount"`
			DeviceID   string `json:"deviceId"`
		} `json:"Parameters"`
	} `json:"Details"`
}

type CheckCountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type merryCount struct {
	DeviceId  string `dynamo:"device_id"`
	CallCount int64  `dynamo:"call_count"`
}

func CheckCount(event CheckCountEvent) (CheckCountResponse, error) {
	db := dynamo.New(session.New(), aws.NewConfig().WithRegion(os.Getenv("Region")))
	merryCountTable := db.Table("merry_count")

	var merryCountData merryCount
	if err := merryCountTable.Get("device_id", event.Details.Parameters.DeviceID).One(&merryCountData); err != nil {
		return CheckCountResponse{Status: "failed", Message: "dynamoDB put error"}, err
	}

	checkCount, err := strconv.ParseInt(event.Details.Parameters.CheckCount, 10, 64)
	if err != nil {
		return CheckCountResponse{Status: "failed", Message: "checkCount is not number"}, errors.New("checkCount is not number")
	}

	if merryCountData.CallCount >= checkCount {
		return CheckCountResponse{Status: "failed", Message: fmt.Sprintf("device:%s count:%d", merryCountData.DeviceId, merryCountData.CallCount)}, errors.New("over count")
	}

	merryCountData.CallCount++
	if checkCount == 3 {
		merryCountData.CallCount = 0
	}
	log.Printf("event %s", event.Details.Parameters.DeviceID)
	if err := merryCountTable.Put(merryCountData).Run(); err != nil {
		return CheckCountResponse{Status: "failed", Message: "table put failed"}, errors.New("merry_count put data failed")
	}

	return CheckCountResponse{Status: "success", Message: fmt.Sprintf("device:%s count:%d", merryCountData.DeviceId, merryCountData.CallCount)}, nil
}

func main() {
	lambda.Start(CheckCount)
}
