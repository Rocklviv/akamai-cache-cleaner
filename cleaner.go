package main

import (
	"os"
	"os/user"
	"fmt"
	"encoding/json"
	"encoding/base64"
	"flag"
	"path/filepath"
	"strings"
	"net/http"
	"bytes"
	"io/ioutil"
)

var (
	// Current version
	VERSION = "snapshot"
	// Directory to read recursively
	dir = ""
	// Domain name if defined
	serverName = ""
	// extensions
	extension = ""
)

type Configuration struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	Domain         string `json:"domain"`
	APIUrl         string `json:"api_url"`
	QueueLengthUri string `json:"queue_length_uri"`
}


type Object struct {
	Objects []string `json:"objects"`
}

type CacheResponse struct {
	EstimatedSeconds 	json.Number	`json:"estimatedSeconds"`
	ProgressUri 		string		`json:"progressUri"`
	PurgeId			string		`json:"purgeId"`
	SupportId		string		`json:"supportId"`
	HttpStatus		json.Number	`json:"httpStatus"`
	Detail			string		`json:"detail"`
	PingAfterSeconds	json.Number	`json:"pingAfterSeconds"`
}

func readConfig() Configuration {
	usr, er := user.Current()
	if er != nil {
		fmt.Println("Cannot get current user: ", er)
	}
	homedir := usr.HomeDir
	file, e := os.Open(homedir + "/.akamai.json")
	if e != nil {
		fmt.Println("Error opening file:", e)
	}
	decoded := json.NewDecoder(file)
	config := Configuration{}
	err := decoded.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config file: ", err)
	}
	return config
}

//
// Reads a directory recursively to collect list of files
// by defined extensions.
//
// Returns array
//
func getListOfFiles(ext string) []string {
	// List of files that will be found by ext.
	var listFiles []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Cannot read directory: ", err)
		}
		if !info.IsDir() {
			extensions := strings.Split(ext, ",")
			httpServers := strings.Split(serverName, ",")
			newpath := strings.Replace(path, dir, "", -1)
			for _, a := range extensions {
				if "." + a == filepath.Ext(info.Name()) {
					for _, b := range httpServers {
						httpPath := b + "/" + newpath
						listFiles = append(listFiles, httpPath)
					}
				}
			}
		}
		return nil
	})
	return listFiles
}

//
// Returns version of application.
//
func returnVersion() {
	fmt.Println(VERSION)
}

func sendRequest() {
	cfg := readConfig()
	fmt.Println("Preparing request to clear akamai for: ", serverName)
	files := getListOfFiles(extension)
	object := Object{Objects: files}
	encoded, err := json.Marshal(object)
	if err != nil {
		fmt.Println("Cannot marshalize value: ", err)
	}
	b := bytes.NewBuffer(encoded)
	fmt.Println("Prepared request: ", b)
	if err != nil {
		fmt.Println("Problem with encoding struct. ERROR: ", err)
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", cfg.APIUrl, b)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Authorization", "Basic " + fmt.Sprintf(basicAuth(cfg.Username, cfg.Password)))
	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Cannot get response or request is not valid.")
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Cannot read response: ", err)
	}
	checkPurge(string(body))
}

func checkPurge(response string) {
	var data CacheResponse
	err := json.Unmarshal([]byte(response), &data)
	if err != nil {
		fmt.Println("Cannot unmarshal data: ", err)
	}
	fmt.Println(data.EstimatedSeconds)
}

func basicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func main() {
	extensionList := flag.String("l", "html", "Define an file extensions that should be cleared from Akamai Cache")
	folder := flag.String("f", "dist/", "Define an folder with files that should be cleared from Akamai Cache")
	domain := flag.String("s", "example.com", "Define a domain that should be used, instead of domain that defined in configuration file.")
	version := flag.String("v", "", "Return version of application.")
	flag.Parse()

	if extensionList != nil {
		extension = *extensionList
	}

	if folder != nil {
		dir = *folder
	}

	if *version != "" {
		returnVersion()
	}

	if domain != nil {
		serverName = *domain
	}

	sendRequest()
}
