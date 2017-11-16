package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"net/http/httputil"
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
		Endpoint string
	}

	Notebook struct {
		Id          int    `json:"-,omitempty"`
		Name        string `json:"name" validate:"required"`
		Description string `json:"description" validate:"required"`
		Protocol    string `json:"protocol" validate:"required"`
		State       string `json:"-" validate:"required"`
	}
)

type (
	NotebookResponse struct {
		Message string         `json:"message"`
		Status  int            `json:"status"`
		Data    []NotebookData `json:"data,omitempty"`
	}

	NotebookData struct {
		Id   int
		Name string
		Ip   string
	}
)

var validate *validator.Validate

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
	settingUpNotebookId(&p.Config)

	//if notebook exists
	if p.Config.Notebook.State == "present" && notebookExists(&p.Config) == false {
		createNotebook(&p.Config)
	} else if p.Config.Notebook.State == "present" {
		Infof("Notebook already present: %s", p.Config.Notebook.Name)
		Infof("Your notebook id: %d", p.Config.Notebook.Id)
	} else if p.Config.Notebook.State == "absent" && notebookExists(&p.Config) == true {
		Infof("Your notebook id: %d", p.Config.Notebook.Id)
		deleteNotebook(&p.Config)
	} else if p.Config.Notebook.State == "absent" {
		Infof("Notebook %s doesn't exists or already deleted, nothing to do ", p.Config.Notebook.Name)
	}

	return nil
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

func settingUpNotebookId(config *Config) {
	url := fmt.Sprintf("%s/notebook", config.Endpoint)
	resp := apiCall(url, "GET", config.Username, config.Password, nil)

	result := NotebookResponse{}
	if resp.StatusCode == 200 {
		err := json.NewDecoder(resp.Body).Decode(&result)

		if err != nil {
			Fatalf("failed to parse /api/v1/notebook to go struct: %+v", resp)
		}
	}

	for _, notebook := range result.Data {
		if notebook.Name == config.Notebook.Name {
			config.Notebook.Id = notebook.Id
		}
	}
}

func deleteNotebook(config *Config) bool {
	Infof("Delete %s notebook\n", config.Notebook.Name)
	url := fmt.Sprintf("%s/notebooks/%d", config.Endpoint, config.Notebook.Id)
	resp := apiCall(url, "DELETE", config.Username, config.Password, nil)

	if resp.StatusCode == 201 {
		Infof("Notebook (%s) will be deleted", config.Notebook.Name)
		return true
	}

	if resp.StatusCode == 404 {
		Errorf("Unable to delete notebook %s", config.Notebook.Name)
		return false
	}

	Fatalf("Unexpected error %+v", resp)
	return false
}

func createNotebook(config *Config) bool {

	Infof("Create %s notebook", config.Notebook.Name)

	url := fmt.Sprintf("%s/notebooks", config.Endpoint)
	param, _ := json.Marshal(config.Notebook)
	resp := apiCall(url, "POST", config.Username, config.Password, bytes.NewBuffer(param))

	if resp.StatusCode == 201 {
		Infof("Notebook (%s) will be installed", config.Notebook.Name)
		return true
	}

	Fatalf("Unexpected error %+v", resp)
	return false
}

func notebookExists(config *Config) bool {
	if config.Notebook.Id > 0 {
		return true
	}
	return false
}
