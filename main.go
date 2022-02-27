package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var uploadFolderID string
var srcFolder string
var fileList []string

func init() {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	uploadFolderID = viper.GetString("UPLOAD")
	srcFolder = viper.GetString("SRC")
	fileList = strings.Split(viper.GetString("FILE_LIST"), ",")
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("./credentials.json")

	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))

	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	//Create folder
	folder := os.Args[1]
	var folderIDList []string
	folderIDList = append(folderIDList, uploadFolderID)

	createFolder, err := srv.Files.Create(&drive.File{Name: folder, MimeType: "application/vnd.google-apps.folder", Parents: folderIDList}).Do()

	fmt.Println(createFolder)
	if err != nil {
		log.Fatalf("Unable to create folder: %v", err)
	}

	var newFolderIDList []string
	newFolderIDList = append(newFolderIDList, createFolder.Id)

	for i, s := range fileList {
		fmt.Println(i, s)
		//Upload file to folder
		uploadToDrive(srcFolder+string(s)+"."+folder+".csv", string(s)+"."+folder+".csv", newFolderIDList, srv)
	}

}

func uploadToDrive(src string, dest string, folderList []string, srv *drive.Service) {
	baseMimeType := "text/csv"                                     // mimeType of file you want to upload
	convertedMimeType := "application/vnd.google-apps.spreadsheet" // mimeType of file you want to convert on Google Drive

	file, err := os.Open(src)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer file.Close()

	f := &drive.File{
		Name:     dest,
		MimeType: convertedMimeType,
		Parents:  folderList,
	}

	res, err := srv.Files.Create(f).Media(file, googleapi.ContentType(baseMimeType)).Do()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf(" %s, %s, %s\n", res.Name, res.Id, res.MimeType)
}
