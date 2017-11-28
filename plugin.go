package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"net/http/httputil"
	"io/ioutil"
	"github.com/oliveagle/jsonpath"
)

type (
	Repo struct {
		Owner   string
		Name    string
		Link    string
		Avatar  string
		Branch  string
		Private bool
		Trusted bool
	}

	Build struct {
		Number   int
		Event    string
		Status   string
		Deploy   string
		Created  int64
		Started  int64
		Finished int64
		Link     string
	}

	Author struct {
		Name   string
		Email  string
		Avatar string
	}

	Commit struct {
		Remote  string
		Sha     string
		Ref     string
		Link    string
		Branch  string
		Message string
		Author  Author
	}

	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}

	Config struct {
		Notebook Notebook
		Username string
		Password string
		Endpoint string `json:"filePath" validate:"required"`
	}

	Notebook struct {
		Id          string `json:"-,omitempty"`
		Name        string `json:"name" validate:"required"`
		FilePath	string `json:"filePath" validate:"required"`
		State       string `json:"filePath" validate:"required"`
	}
)

var validate *validator.Validate

func notebookExists(config *Config) bool {
	if config.Notebook.Id != "" {
		return true
	}
	return false
}

func apiCall(url string, method string, username string, password string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, url, body)

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		Fatalf("failed to build http request: %v", err)
	}

	req.SetBasicAuth(username, password)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		Fatalf("failed to call \"%s\" on %s: %+v", method, url, err)
	}

	debugReq, _ := httputil.DumpRequest(req, true)
	Debugf("Request %s", debugReq)
	debugResp, _ := httputil.DumpResponse(resp, true)
	Debugf("Response %s", debugResp)

	defer resp.Body.Close()
	return resp
}

func lookupNotebookId(config *Config) {
	url := fmt.Sprintf("%s/api/notebook", config.Endpoint)
	resp := apiCall(url, "GET", config.Username, config.Password, nil)

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
	if resp.StatusCode == 200 {
		err := json.NewDecoder(resp.Body).Decode(&result)

		if err != nil {
			Fatalf("failed to parse /api/v1/notebook to go struct: %+v", resp)
		}
	}

	for _, notebook := range result.Data {
		if notebook.Name == config.Notebook.Name {
			config.Notebook.Id = notebook.Id
			Infof("Notebook %s found with id: %s", config.Notebook.Name, config.Notebook.Id)
		}
	}

	if config.Notebook.Id == "" {
		Infof("Notebook not found with name: %s", config.Notebook.Name)
	}
}

func deleteNotebook(config *Config) bool {
	Infof("Deleting notebook ...")
	url := fmt.Sprintf("%s/api/notebook/%s", config.Endpoint, config.Notebook.Id)
	resp := apiCall(url, "DELETE", config.Username, config.Password, nil)

	if resp.StatusCode == 200 {
		Infof("Notebook %s (%s) has been deleted", config.Notebook.Name, config.Notebook.Id)
		config.Notebook.Id = ""
		return true
	}

	if resp.StatusCode == 404 {
		Errorf("Unable to delete notebook %s", config.Notebook.Name)
		return false
	}

	Fatalf("Unexpected error %+v", resp)
	return false
}

func importNotebook(config *Config) bool {

	Infof("Importing notebook ...")
	notebookData, err := ioutil.ReadFile(config.Notebook.FilePath)
	if err != nil {
		panic(err)
	}

	url := fmt.Sprintf("%s/api/notebook/import", config.Endpoint)
	resp := apiCall(url, "POST", config.Username, config.Password, bytes.NewBuffer(notebookData))

	if resp.StatusCode == 200 {
		//var result = map[string] string{}
		result := map[string] string{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			Fatalf("failed to parse /api/notebook/import to go struct: %+v", err)
		}
		config.Notebook.Id = result["body"]
		Infof("Notebook %s has been imported with id: %s", config.Notebook.Name, config.Notebook.Id)
		return true
	}

	Fatalf("Unexpected error %+v", resp)
	return false
}

func createNotebook(p *Plugin) bool {

	if notebookExists(&p.Config) == false {
		importNotebook(&p.Config)
	} else {
		if notebookInProgress(&p.Config) == false {
			Infof("Notebook %s (%s) already exists with same name, will be recreated",
				p.Config.Notebook.Name, p.Config.Notebook.Id)
			deleteNotebook(&p.Config)
			importNotebook(&p.Config)
		} else {
			Infof("Notebook %s (%s) already exists with same name and is in progress",
				p.Config.Notebook.Name, p.Config.Notebook.Id)
			return false;
		}
	}
	return true;

}

func runNotebook(config *Config) bool {

	Infof("Running notebook ...")
	url := fmt.Sprintf("%s/api/notebook/job/%s?waitToFinish=false", config.Endpoint, config.Notebook.Id)
	resp := apiCall(url, "POST", config.Username, config.Password, nil)

	if resp.StatusCode == 200 {
		Infof("Notebook %s (%s) has been started", config.Notebook.Name, config.Notebook.Id)
		return true
	}
	Fatalf("Unexpected error %+v", resp)
	return false
}

func notebookInProgress(config *Config) bool {

	Infof("Checking notebook status ...")
	url := fmt.Sprintf("%s/api/notebook/job/%s", config.Endpoint, config.Notebook.Id)
	resp := apiCall(url, "GET", config.Username, config.Password, nil)
	if resp.StatusCode != 200 {
		return false
	}
	data, err := ioutil.ReadAll(resp.Body)
	var jsonData interface{}
	json.Unmarshal(data, &jsonData)

	res, err := jsonpath.JsonPathLookup(jsonData, "$.body[?(@.progress < 100)].status")
	if err != nil {
		Fatalf("failed to lookup path in json: %s", err)
	}

	var inProgress = false
	if len(res.([]interface{})) > 0 {
		inProgress = true
	}
	Infof("Notebook %s (%s) is in progress: %t", config.Notebook.Name, config.Notebook.Id, inProgress)
	return inProgress
}

func (p *Plugin) Exec() error {
	validate = validator.New()

	err := validate.Struct(p)
	if err != nil {
		for _, v := range err.(validator.ValidationErrors) {
			Errorf("[%s] field validation error (%+v)", v.Field(), v)
		}
		return nil
	}

	Infof("Notebook desired state: %s", p.Config.Notebook.State)
	lookupNotebookId(&p.Config)

	switch p.Config.Notebook.State {
	case "created" :
		createNotebook(p)
	case "running" :
		// run notebook if has been successfully created
		if createNotebook(p) == true {
			runNotebook(&p.Config)
		}
	case "deleted"  :
		if notebookExists(&p.Config) == true {
			deleteNotebook(&p.Config)
		} else {
			Infof("Notebook %s doesn't exists, nothing to do ", p.Config.Notebook.Name)
		}
	}

	return nil
}
