package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"workflow-engine/internal/model"

	"github.com/camunda/zeebe/clients/go/v8/pkg/entities"
	"github.com/camunda/zeebe/clients/go/v8/pkg/worker"
	"github.com/camunda/zeebe/clients/go/v8/pkg/zbc"
)

const ZeebeAddr = "0.0.0.0:26500"

var readyClose = make(chan struct{})

func main() {
	gatewayAddr := os.Getenv("ZEEBE_ADDRESS")
	plainText := false

	if gatewayAddr == "" {
		gatewayAddr = ZeebeAddr
		plainText = true
	}

	zbClient, err := zbc.NewClient(&zbc.ClientConfig{
		GatewayAddress:         gatewayAddr,
		UsePlaintextConnection: plainText,
	})

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// create a new process instance
	variables := make(map[string]interface{})
	variables["orderId"] = "31243"

	request, err := zbClient.NewCreateInstanceCommand().BPMNProcessId("fww-bpm").LatestVersion().VariablesFromMap(variables)
	if err != nil {
		panic(err)
	}

	result, err := request.Send(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println(result.String())

	go zbClient.NewJobWorker().JobType("check-blacklist").Handler(handleCheckBlacklist).Open()
	go zbClient.NewJobWorker().JobType("check-dukcapil").Handler(handleCheckDukcapil).Open()
	go zbClient.NewJobWorker().JobType("check-pedulilindungi").Handler(handleCheckPeduliLindungi).Open()
	go zbClient.NewJobWorker().JobType("send-email-booking").Handler(handleSendEmailBooking).Open()
	go zbClient.NewJobWorker().JobType("send-email-unpaid").Handler(handleSendEmailUnpaid).Open()

	// res, err := zbClient.NewCompleteJobCommand().JobKey(4503599628408446).Send(ctx)
	// fmt.Println("Complete user task", res, err)
	http.HandleFunc("/start", createStartHandler(zbClient))
	err = http.ListenAndServe(":3003", nil)
	if err != nil {
		panic(err)
	}

}

type BoundHandler func(w http.ResponseWriter, r *http.Request)

type InitialVariables struct {
	Name string `json:"name"`
}

func createStartHandler(client zbc.Client) BoundHandler {
	f := func(w http.ResponseWriter, r *http.Request) {
		variables := &InitialVariables{"Syamsul"}
		ctx := context.Background()
		request, err := client.NewCreateInstanceCommand().BPMNProcessId("fww-reservation").LatestVersion().VariablesFromObject(variables)
		if err != nil {
			panic(err)
		}
		response, _ := request.WithResult().Send(ctx)
		var result map[string]interface{}
		json.Unmarshal([]byte(response.Variables), &result)
		fmt.Fprint(w, result["say"])
	}
	return f
}

func handleEmailBooking(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	variables["blacklistUser"] = false
	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully user blacklist completed job")
}

func handleCheckBlacklist(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	blakclistRegulationApiURL := "http://localhost:3004/check/blacklist"

	passengerId, _ := variables["passengerId"].(string)
	userId := model.ReqKtp{
		KTP: passengerId,
	}

	payload, _ := json.Marshal(userId)

	response, err := http.Post(blakclistRegulationApiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println("Error http request:", err)
	}

	var responseBody []byte
	if response.Body != nil {
		responseBody, err = io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error read body:", err)
			return
		}
	}
	var blacklistResp model.RegulationtResponse
	err = json.Unmarshal(responseBody, &blacklistResp)
	if err != nil {
		fmt.Println("Error unmarshal body:", err)
		return
	}
	variables["blacklistUser"] = blacklistResp.Status
	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully user blacklist completed job")
}

func handleCheckDukcapil(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	dukcapilRegulationApiURL := "http://localhost:3004/check/dukcapil"

	passengerId, _ := variables["passengerId"].(string)
	userId := model.ReqKtp{
		KTP: passengerId,
	}

	payload, _ := json.Marshal(userId)

	response, err := http.Post(dukcapilRegulationApiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println("Error http request:", err)
	}

	var responseBody []byte
	if response.Body != nil {
		responseBody, err = io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error read body:", err)
			return
		}
	}
	var blacklistResp model.RegulationtResponseString
	err = json.Unmarshal(responseBody, &blacklistResp)
	if err != nil {
		fmt.Println("Error unmarshal body:", err)
		return
	}
	variables["dukcapil"] = blacklistResp.Status
	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully check dukcapil completed job")
}

func handleCheckPeduliLindungi(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}

	pedulilindungiRegulationApiURL := "http://localhost:3004/check/pedulilindungi"

	passengerId, _ := variables["passengerId"].(string)
	userId := model.ReqKtp{
		KTP: passengerId,
	}

	payload, _ := json.Marshal(userId)

	response, err := http.Post(pedulilindungiRegulationApiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println("Error http request:", err)
	}

	var responseBody []byte
	if response.Body != nil {
		responseBody, err = io.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error read body:", err)
			return
		}
	}
	var blacklistResp model.RegulationtResponseString
	err = json.Unmarshal(responseBody, &blacklistResp)
	if err != nil {
		fmt.Println("Error unmarshal body:", err)
		return
	}
	variables["peduliLindungi"] = blacklistResp.Status
	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully check peduli lindungi completed job")

	// close(readyClose)
}

func handleSendEmailBooking(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}
	fmt.Println(variables)
	sendEmailURL := "http://localhost:3002/send-email"

	reservationId, _ := variables["reservationId"].(float64)
	userId := model.ReqEmail{
		ReservationId: int(reservationId),
	}
	fmt.Println("reservation ID: ", variables["reservationId"].(float64))
	payload, _ := json.Marshal(userId)

	response, err := http.Post(sendEmailURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println("Error http request:", err, response.StatusCode)
		return
	}
	fmt.Println("Succesfully send email reservation")
	variables["sendEmailReservation"] = "success"

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully check send email reservation completed job")

	// close(readyClose)
}

// send email unpaid
func handleSendEmailUnpaid(client worker.JobClient, job entities.Job) {
	jobKey := job.GetKey()

	variables, err := job.GetVariablesAsMap()
	if err != nil {
		// failed to handle job as we require the variables
		failJob(client, job)
		return
	}
	fmt.Println(variables)
	sendEmailURL := "http://localhost:3002/send-email"

	reservationId, _ := variables["reservationId"].(float64)
	userId := model.ReqEmail{
		ReservationId: int(reservationId),
	}
	fmt.Println("reservation ID: ", variables["reservationId"].(float64))
	payload, _ := json.Marshal(userId)

	response, err := http.Post(sendEmailURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println("Error http request:", err, response.StatusCode)
		return
	}
	fmt.Println("Succesfully send email reservation")
	variables["sendEmail"] = "success"

	request, err := client.NewCompleteJobCommand().JobKey(jobKey).VariablesFromMap(variables)
	if err != nil {
		// failed to set the updated variables
		failJob(client, job)
		return
	}

	log.Println("Complete job", jobKey, "of type", job.Type)

	ctx := context.Background()
	_, err = request.Send(ctx)
	if err != nil {
		panic(err)
	}

	log.Println("Successfully check send email reservation completed job")

	// close(readyClose)
}

func failJob(client worker.JobClient, job entities.Job) {
	log.Println("Failed to complete job", job.GetKey())

	ctx := context.Background()
	_, err := client.NewFailJobCommand().JobKey(job.GetKey()).Retries(job.Retries - 1).Send(ctx)
	if err != nil {
		panic(err)
	}
}
