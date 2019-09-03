package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mohae/struct2csv"
	"golang.org/x/crypto/ssh/terminal"
)

var authResponse authenticateResponseType

func main() {

	var userName, password string
	userName = "awswaldenproductionaccount@laureate.net"
	password = "<password>"

	login(userName, password)
	listResp := listTimeLines()
	fmt.Printf("Timelines: %v\n", len(listResp))
	var tableRows []relTable
	for _, rel := range listResp {
		if rel.Status != "ACTIVE" && rel.Relationship.Activity.ExecutionTime > 0 {
			tableRows = append(tableRows,
				relTable{
					Status:    rel.Status,
					Source:    rel.Relationship.Source.Nfs.Host + ":" + rel.Relationship.Source.Nfs.Export,
					Target:    rel.Relationship.Target.Nfs.Host + ":" + rel.Relationship.Target.Nfs.Export,
					StartTime: fmt.Sprint("", rel.Relationship.Activity.StartTime.Format("02/01/2006 15:04:05")),
					StopTime:  fmt.Sprint("", rel.Relationship.Activity.EndTime.Format("02/01/2006 15:04:05")),
					Duration:  fmtDuration(rel.Relationship.Activity.EndTime.Sub(rel.Relationship.Activity.StartTime)),
					CountS:    strconv.FormatInt(rel.Relationship.Activity.FilesCopied, 10),
					SizeS:     strconv.FormatInt(rel.Relationship.Activity.BytesCopied/1024/1024, 10),
				})
		} else {

		}
	}
	enc := struct2csv.New()
	rows, _ := enc.Marshal(tableRows)
	csvfile, err := os.Create("output.csv")

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvwriter := csv.NewWriter(csvfile)

	for _, row := range rows {
		_ = csvwriter.Write(row)
	}

	csvwriter.Flush()

	csvfile.Close()
}
func listTimeLines() timelineType {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://cloudsync.netapp.com/api/timelines-v2", nil)
	req.Header.Add("Authorization", "Bearer "+authResponse.AccessToken)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	var relResponeV2 timelineType
	json.Unmarshal(body, &relResponeV2)
	return relResponeV2
}
func login(username string, password string) {
	Authenticate := map[string]interface{}{
		"username":   username,
		"scope":      "profile",
		"audience":   "https://api.cloud.netapp.com",
		"client_id":  "UaVhOIXMWQs5i1WdDxauXe5Mqkb34NJQ",
		"grant_type": "password",
		"password":   password,
	}

	bytesRepresentation, err := json.Marshal(Authenticate)
	if err != nil {
		log.Fatalln(err)
	}
	resp, err := http.Post("https://netapp-cloud-account.auth0.com/oauth/token", "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil || resp.StatusCode != 200 {
		log.Fatalln(resp)
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	json.Unmarshal(body, &authResponse)
}
func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		fmt.Println("\nPassword typed: " + string(bytePassword))
	}
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	return fmt.Sprintf("%02d:%02d", h, m)
}

type authenticateResponseType struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
type relTable struct {
	Status    string `csv:"Status"`
	Source    string `csv:"Source"`
	Target    string `csv:"Target"`
	StartTime string `csv:"StartTime"`
	StopTime  string `csv:"StopTime"`
	Duration  string `csv:"Duration"`
	CountS    string `csv:"Count"`
	SizeS     string `csv:"Size"`
}
type timelineType []struct {
	Status       string `json:"status"`
	RequestID    string `json:"requestId"`
	Relationship struct {
		IsCvo  bool `json:"isCvo"`
		Source struct {
			Protocol string `json:"protocol"`
			Nfs      struct {
				Host     string `json:"host"`
				Export   string `json:"export"`
				Path     string `json:"path"`
				Version  string `json:"version"`
				Provider string `json:"provider"`
			} `json:"nfs"`
		} `json:"source"`
		Target struct {
			Protocol string `json:"protocol"`
			Nfs      struct {
				Host     string `json:"host"`
				Export   string `json:"export"`
				Path     string `json:"path"`
				Version  string `json:"version"`
				Provider string `json:"provider"`
			} `json:"nfs"`
		} `json:"target"`
		IsQstack       bool   `json:"isQstack"`
		RelationshipID string `json:"relationshipId"`
		Group          string `json:"group"`
		DataBrokerID   string `json:"dataBrokerId"`
		Activity       struct {
			Type                 string    `json:"type"`
			Status               string    `json:"status"`
			FailureMessage       string    `json:"failureMessage"`
			ExecutionTime        int64     `json:"executionTime"`
			StartTime            time.Time `json:"startTime"`
			EndTime              time.Time `json:"endTime"`
			BytesMarkedForCopy   int64     `json:"bytesMarkedForCopy"`
			FilesMarkedForCopy   int64     `json:"filesMarkedForCopy"`
			DirsMarkedForCopy    int64     `json:"dirsMarkedForCopy"`
			FilesCopied          int64     `json:"filesCopied"`
			BytesCopied          int64     `json:"bytesCopied"`
			DirsCopied           int64     `json:"dirsCopied"`
			FilesFailed          int64     `json:"filesFailed"`
			BytesFailed          int64     `json:"bytesFailed"`
			DirsFailed           int64     `json:"dirsFailed"`
			FilesMarkedforRemove int64     `json:"filesMarkedforRemove"`
			BytesMarkedForRemove int64     `json:"bytesMarkedForRemove"`
			DirsMarkedForRemove  int64     `json:"dirsMarkedForRemove"`
			FilesRemoved         int64     `json:"filesRemoved"`
			BytesRemoved         int64     `json:"bytesRemoved"`
			DirsRemoved          int64     `json:"dirsRemoved"`
			BytesRemovedFailed   int64     `json:"bytesRemovedFailed"`
			FilesRemovedFailed   int64     `json:"filesRemovedFailed"`
			FilesMarkedForGrace  int64     `json:"filesMarkedForGrace"`
			BytesMarkedForGrace  int64     `json:"bytesMarkedForGrace"`
			DirsMarkedForGrace   int64     `json:"dirsMarkedForGrace"`
			FilesMarkedForIgnore int64     `json:"filesMarkedForIgnore"`
			DirsScanned          int64     `json:"dirsScanned"`
			FilesScanned         int64     `json:"filesScanned"`
			DirsFailedToScan     int64     `json:"dirsFailedToScan"`
			BytesScanned         int64     `json:"bytesScanned"`
			Progress             int64     `json:"progress"`
			LastMessageTime      time.Time `json:"lastMessageTime"`
		} `json:"activity"`
	} `json:"relationship,omitempty"`
	DataBroker struct {
		Name         string `json:"name"`
		DataBrokerID string `json:"dataBrokerId"`
	} `json:"dataBroker,omitempty"`
	Group struct {
		Name    string `json:"name"`
		GroupID string `json:"groupId"`
	} `json:"group,omitempty"`
	Summary        string `json:"summary"`
	CreatedAt      int64  `json:"createdAt"`
	FailureMessage string `json:"failureMessage,omitempty"`
	ID             string `json:"id"`
}

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

func secondsToHuman(input int) (result string) {
	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if months > 0 {
		result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if weeks > 0 {
		result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if days > 0 {
		result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if hours > 0 {
		result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if minutes > 0 {
		result = plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else {
		result = plural(int(seconds), "second")
	}

	return
}
