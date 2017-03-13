package main

import (
	"os"
	"os/user"
	"fmt"
	"encoding/json"
	"flag"
	"path/filepath"
	"strings"
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

func readConfig() {
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
}

//
// Reads a directory recursively to collect list of files
// by defined extensions.
//
// Returns array
//
func getListOfFiles(ext string) []string {
	fmt.Println("Extension: ", ext)
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
				if "."+a == filepath.Ext(info.Name()) {
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

func returnVersion() {
	fmt.Println(VERSION)
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
		fmt.Println(serverName)
	}

	readConfig()
	files := getListOfFiles(*extensionList)
	fmt.Println("List of files: ", files)
}
