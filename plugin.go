package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/oliveagle/jsonpath"
	log "github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

type (
	Plugin struct {
		Config Config
	}

	Config struct {
		Notebook Notebook
		Username string
		Password string
		Endpoint string `json:"endpoint" validate:"required"`
	}

	Notebook struct {
		Id       string `json:"-,omitempty"`
		Name     string `json:"name" validate:"required"`
		FilePath string `json:"filePath" validate:"required"`
		State    string `json:"state" validate:"required"`
	}
)

var validate *validator.Validate

func notebookExists(config *Config) bool {
	if config.Notebook.Id != "" {
		return true
	}
	return false
}

func apiCall(url string, method string, username string, password string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Fatalf("could not create request. method: [%s], url: [%s]", method, url)
	}

	if method == http.MethodPost {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		log.Fatalf("failed to build http request: %v", err)
	}

	req.SetBasicAuth(username, password)
	req.Header.Add("Accept", "application/json")

	var netClient = &http.Client{
		//Timeout: time.Second * 1,
	}
	resp, err := netClient.Do(req)

	if err != nil {
		log.Errorf("failed to call \"%s\" on %s: %+v", method, url, err)
		return nil, err
	}

	debugReq, _ := httputil.DumpRequest(req, true)
	log.Debugf("Request %s", debugReq)
	debugResp, _ := httputil.DumpResponse(resp, true)
	log.Debugf("Response %s", debugResp)

	defer resp.Body.Close()
	return resp, nil
}

func lookupNotebookId(config *Config) {
	url := fmt.Sprintf("%s/api/notebook", config.Endpoint)

	resp, err := apiCall(url, http.MethodGet, config.Username, config.Password, nil)
	for i := 0; i < 10 && err != nil; i++ {
		time.Sleep(5 * time.Second)
		log.Infof("retry api call %d time(s)", i + 1)
		resp, err = apiCall(url, http.MethodGet, config.Username, config.Password, nil)
	}

	if err != nil {
		log.Fatalf("failed to lookup notebook %+v", err)
	}

	type (
		NotebookData struct {
			Id   string
			Name string
		}

		GetNotebooksResponse struct {
			Message string         `json:"message"`
			Status  string         `json:"status"`
			Data    []NotebookData `json:"body,omitempty"`
		}
	)

	result := GetNotebooksResponse{}
	if resp.StatusCode == http.StatusOK {
		err := json.NewDecoder(resp.Body).Decode(&result)

		if err != nil {
			log.Fatalf("failed to parse /api/v1/notebook to go struct: %+v", resp)
		}
	}

	for _, notebook := range result.Data {
		if notebook.Name == config.Notebook.Name {
			config.Notebook.Id = notebook.Id
			log.Infof("Notebook %s found with id: %s", config.Notebook.Name, config.Notebook.Id)
		}
	}

	if config.Notebook.Id == "" {
		log.Infof("Notebook not found with name: %s", config.Notebook.Name)
	}
}

func deleteNotebook(config *Config) bool {
	log.Infof("deleting notebook ...")
	url := fmt.Sprintf("%s/api/notebook/%s", config.Endpoint, config.Notebook.Id)
	resp, _ := apiCall(url, http.MethodDelete, config.Username, config.Password, nil)

	if resp.StatusCode == http.StatusOK {
		log.Infof("Notebook %s (%s) has been deleted", config.Notebook.Name, config.Notebook.Id)
		config.Notebook.Id = ""
		return true
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Errorf("Unable to delete notebook %s", config.Notebook.Name)
		return false
	}

	log.Fatalf("Unexpected error %+v", resp)
	return false
}

func importNotebook(config *Config) bool {

	log.Infof("Importing notebook ...")
	notebookData, err := ioutil.ReadFile(config.Notebook.FilePath)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s/api/notebook/import", config.Endpoint)
	resp, _ := apiCall(url, http.MethodPost, config.Username, config.Password, bytes.NewBuffer(notebookData))

	if resp.StatusCode == http.StatusOK {
		//var result = map[string] string{}
		result := map[string]string{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			log.Fatalf("failed to parse /api/notebook/import to go struct: %+v", err)
		}
		config.Notebook.Id = result["body"]
		log.Infof("Notebook %s has been imported with id: %s", config.Notebook.Name, config.Notebook.Id)
		return true
	}

	log.Fatalf("Unexpected error %+v", resp)
	return false
}

func createNotebook(p *Plugin) bool {

	if notebookExists(&p.Config) == false {
		importNotebook(&p.Config)
	} else {
		if notebookInProgress(&p.Config) == false {
			log.Infof("Notebook %s (%s) already exists with same name, will be recreated",
				p.Config.Notebook.Name, p.Config.Notebook.Id)
			deleteNotebook(&p.Config)
			importNotebook(&p.Config)
		} else {
			log.Infof("Notebook %s (%s) already exists with same name and is in progress",
				p.Config.Notebook.Name, p.Config.Notebook.Id)
			return false
		}
	}
	return true

}

func runNotebook(config *Config) bool {

	log.Infof("Running notebook ...")
	url := fmt.Sprintf("%s/api/notebook/job/%s?waitToFinish=false", config.Endpoint, config.Notebook.Id)
	resp, _ := apiCall(url, http.MethodPost, config.Username, config.Password, nil)

	if resp.StatusCode == http.StatusOK {
		log.Infof("Notebook %s (%s) has been started", config.Notebook.Name, config.Notebook.Id)
		return true
	}
	log.Fatalf("Unexpected error %+v", resp)
	return false
}

func notebookInProgress(config *Config) bool {

	log.Infof("Checking notebook status ...")
	url := fmt.Sprintf("%s/api/notebook/job/%s", config.Endpoint, config.Notebook.Id)
	resp, _ := apiCall(url, http.MethodGet, config.Username, config.Password, nil)
	if resp.StatusCode != http.StatusOK {
		return false
	}
	data, err := ioutil.ReadAll(resp.Body)
	var jsonData interface{}
	json.Unmarshal(data, &jsonData)

	res, err := jsonpath.JsonPathLookup(jsonData, "$.body[?(@.progress < 100)].status")
	if err != nil {
		log.Fatalf("failed to lookup path in json: %s", err)
	}

	var inProgress = false
	if len(res.([]interface{})) > 0 {
		inProgress = true
	}
	log.Infof("Notebook %s (%s) is in progress: %t", config.Notebook.Name, config.Notebook.Id, inProgress)
	return inProgress
}

func (p *Plugin) Exec() error {
	validate = validator.New()

	err := validate.Struct(p)
	if err != nil {
		for _, v := range err.(validator.ValidationErrors) {
			log.Errorf("[%s] field validation error (%+v)", v.Field(), v)
		}
		return nil
	}

	log.Infof("Notebook desired state: %s", p.Config.Notebook.State)
	lookupNotebookId(&p.Config)

	switch p.Config.Notebook.State {
	case "present":
		createNotebook(p)
	case "running":
		// run notebook if has been successfully created
		if createNotebook(p) == true {
			runNotebook(&p.Config)
		}
	case "absent":
		if notebookExists(&p.Config) == true {
			deleteNotebook(&p.Config)
		} else {
			log.Infof("Notebook [%s] doesn't exists, nothing to do ", p.Config.Notebook.Name)
		}
	}

	return nil
}
