package adapter

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/AsGz/httpAuthClient"
	"github.com/aymerick/raymond"
	"github.com/nu7hatch/gouuid"
)

type OMClient struct {
	Url      string
	Username string
	ApiKey   string
}

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AgentAPIKey string `json:"agentApiKey"`
}

func (oc OMClient) LoadDoc(key string, ctx map[string]string) (string, error) {
	docs := map[string]string{
		"standalone":        "om_cluster_docs/standalone.json",
		"single_replica_set": "om_cluster_docs/replica_set.json",
		"sharded_cluster":    "om_cluster_docs/3_2_cluster.json",
	}

	raymond.RegisterHelper("password", func() string {
		return oc.RandomString(32)
	})

	path := docs[key]
	asset, _ := Asset(path)

	tpl := string(asset)
	result, err := raymond.Render(tpl, ctx)

	if err != nil {
		return "", nil
	}

	return result, nil
}

func (oc OMClient) CreateGroup() (Group, error) {

	u, err := uuid.NewV4()
	groupName := fmt.Sprintf("pcf_%s", u)
	body := strings.NewReader(fmt.Sprintf("{\"name\": \"%s\"}", groupName))

	var group Group

	resp, err := oc.doRequest("POST", "/api/public/v1.0/groups", body)

	if err != nil {
		return group, err
	}

	var b []byte
	b, err = ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, &group)

	return group, nil
}

func (oc OMClient) ConfigureGroup(configurationDoc string, groupId string) error {

	url := fmt.Sprintf("/api/public/v1.0/groups/%s/automationConfig", groupId)
	body := strings.NewReader(configurationDoc)

	resp, err := oc.doRequest("PUT", url, body)
	var b []byte
	b, err = ioutil.ReadAll(resp.Body)

	log.Println(string(b))

	if err != nil {
		return err
	}

	return nil
}

func (oc OMClient) doRequest(method string, path string, body io.Reader) (*http.Response, error) {

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", oc.Url, path), body)
	req.Header.Set("Content-Type", "application/json")

	err = httpAuthClient.ApplyHttpDigestAuth(oc.Username, oc.ApiKey, fmt.Sprintf("%s%s", oc.Url, path), req)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalf("could not post: %v", err)
		return nil, err
	}

	return resp, nil
}

func (oc OMClient) RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// func (oc OMClient) PostDoc(url string, username string, apiKey string) {
//
// }