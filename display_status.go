package main

import (
	"bufio"
	"bytes"
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

	"github.com/kataras/tablewriter"
	"github.com/landoop/tableprinter"
	"golang.org/x/crypto/ssh/terminal"
)

var authResponse AuthenticateResponseType

func main() {

	var userName, password string
	userName = ""
	password = ""

	login(userName, password)
	printer := tableprinter.New(os.Stdout)
	printer.BorderTop, printer.BorderBottom, printer.BorderLeft, printer.BorderRight = true, true, true, true
	printer.CenterSeparator = "│"
	printer.ColumnSeparator = "│"
	printer.RowSeparator = "─"
	printer.HeaderBgColor = tablewriter.BgBlackColor
	printer.HeaderFgColor = tablewriter.FgGreenColor
	for {
		listResp := listRelationships()
		fmt.Printf("Relationships: %v\n", len(listResp))
		var tableRows []relTable
		for _, rel := range listResp {
			tableRows = append(tableRows,
				relTable{
					Source:    rel.Source.Nfs.Host + ":" + rel.Source.Nfs.Export,
					Target:    rel.Target.Nfs.Host + ":" + rel.Target.Nfs.Export,
					StartTime: rel.Activity.StartTime,
					StopTime:  time.Unix(rel.Activity.EndTime, 0),
					Status:    rel.Activity.Status + " (" + strconv.FormatInt(rel.Activity.ExecutionTime, 10) + ")",
					CountS:    strconv.FormatInt(rel.Activity.FilesCopied, 10) + " out of " + strconv.FormatInt(rel.Activity.FilesMarkedForCopy, 10),
					SizeS:     strconv.FormatInt(rel.Activity.BytesCopied, 10) + " out of " + strconv.FormatInt(rel.Activity.BytesCopied, 10),
				})
		}
		printer.Print(tableRows)
		time.Sleep(time.Second * 10)
	}
	os.Exit(0)
}
func listRelationships() (relResponeV2 ListV2Type) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://cloudsync.netapp.com/api/relationships-v2", nil)
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
	if err != nil {
		log.Fatalln(err)
	}
	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	json.Unmarshal(body, &authResponse)
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

type AuthenticateResponseType struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}
type relTable struct {
	Source    string    `header:"Source"`
	Target    string    `header:"Target"`
	Status    string    `header:"Status"`
	StartTime time.Time `header:"StartTime"`
	StopTime  time.Time `header:"StopTime"`
	Duration  time.Time `header:"Duration"`
	CountS    string    `header:"Count"`
	SizeS     string    `header:"Size"`
}
type ListV2Type []struct {
	IsQstack bool   `json:"isQstack"`
	IsCvo    bool   `json:"isCvo"`
	Phase    string `json:"phase"`
	Source   struct {
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
	Sync struct {
		InDays    int       `json:"inDays"`
		InHours   int       `json:"inHours"`
		InMinutes int       `json:"inMinutes"`
		AutoSync  string    `json:"autoSync"`
		NextTime  time.Time `json:"nextTime"`
	} `json:"sync"`
	Settings struct {
		GracePeriod    int  `json:"gracePeriod"`
		DeleteOnTarget bool `json:"deleteOnTarget"`
		Retries        int  `json:"retries"`
		Files          struct {
			ExcludeExtensions []interface{} `json:"excludeExtensions"`
			MaxSize           int64         `json:"maxSize"`
			MinSize           int           `json:"minSize"`
			MinDate           string        `json:"minDate"`
			MaxDate           interface{}   `json:"maxDate"`
		} `json:"files"`
		FileTypes struct {
			Files       bool `json:"files"`
			Directories bool `json:"directories"`
			Symlinks    bool `json:"symlinks"`
		} `json:"fileTypes"`
	} `json:"settings"`
	DataBroker struct {
		LastPing struct {
			Wasabi int64 `json:"wasabi"`
		} `json:"lastPing"`
		Type             string  `json:"type"`
		Name             string  `json:"name"`
		GroupID          string  `json:"groupId"`
		CreatedAt        int64   `json:"createdAt"`
		TransferRate     float64 `json:"transferRate"`
		UpdateNewVersion bool    `json:"updateNewVersion"`
		ID               string  `json:"id"`
		Placement        struct {
			Hostname      string `json:"hostname"`
			Platform      string `json:"platform"`
			PrivateIP     string `json:"privateIp"`
			Version       string `json:"version"`
			Os            string `json:"os"`
			Release       string `json:"release"`
			OsTotalMem    string `json:"osTotalMem"`
			Node          string `json:"node"`
			Cpus          string `json:"cpus"`
			ProcessMaxMem string `json:"processMaxMem"`
		} `json:"placement"`
		Status   string `json:"status"`
		FileLink string `json:"fileLink"`
	} `json:"dataBroker"`
	Group struct {
		DataBrokers []struct {
			LastPing struct {
				Wasabi int64 `json:"wasabi"`
			} `json:"lastPing"`
			Type             string  `json:"type"`
			Name             string  `json:"name"`
			GroupID          string  `json:"groupId"`
			CreatedAt        int64   `json:"createdAt"`
			TransferRate     float64 `json:"transferRate"`
			UpdateNewVersion bool    `json:"updateNewVersion"`
			ID               string  `json:"id"`
			Placement        struct {
				Hostname      string `json:"hostname"`
				Platform      string `json:"platform"`
				PrivateIP     string `json:"privateIp"`
				Version       string `json:"version"`
				Os            string `json:"os"`
				Release       string `json:"release"`
				OsTotalMem    string `json:"osTotalMem"`
				Node          string `json:"node"`
				Cpus          string `json:"cpus"`
				ProcessMaxMem string `json:"processMaxMem"`
			} `json:"placement"`
			Status   string `json:"status"`
			FileLink string `json:"fileLink"`
			Message  string `json:"message,omitempty"`
		} `json:"dataBrokers"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"createdAt"`
		ID        string    `json:"id"`
	} `json:"group"`
	StartTime        time.Time `json:"startTime"`
	CreatedAt        int64     `json:"createdAt"`
	ScannerQueue     string    `json:"scannerQueue,omitempty"`
	TransferrerQueue string    `json:"transferrerQueue,omitempty"`
	AutoSync         string    `json:"autoSync"`
	ID               string    `json:"id"`
	RelationshipID   string    `json:"relationshipId"`
	Activity         struct {
		Type                 string    `json:"type"`
		Status               string    `json:"status"`
		FailureMessage       string    `json:"failureMessage"`
		ExecutionTime        int64     `json:"executionTime"`
		StartTime            time.Time `json:"startTime"`
		EndTime              int64     `json:"endTime"`
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
	EndTime time.Time `json:"endTime,omitempty"`
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
