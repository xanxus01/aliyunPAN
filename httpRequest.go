package aliyunDriver

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type response struct {
	StatusCode int
	Status string
	Data []byte
}

func (d Driver) get(url string, body io.Reader) (response *http.Response, err error) {
	req, err := d.getRequest("GET", url, body)

	if err != nil {
		return
	}

	response, err = http.DefaultClient.Do(req)
	return
}

func (d Driver) post(url string, body io.Reader, res interface{}) (*response,error) {
	req, err := d.getRequest("POST", url, body)

	if err != nil {
		return nil,err
	}

	return doAndParse(req, res)
}

func (d Driver) postJson(url string, body io.Reader, res interface{}) (*response,error) {
	req, err := d.getRequest("POST", url, body)

	if err != nil {
		return nil,err
	}
	req.Header.Set("content-type","application/json;charset=UTF-8")

	return doAndParse(req, res)
}

func authPost(url string, body io.Reader, res interface{}) (response *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return
	}

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	data := make([]byte, 0)
	data,err = parseRes(response.Body, res)
	if err != nil {
		log("err", url, err.Error(), data)
	}

	return
}

func authGet(url string, body io.Reader, res interface{}) (response *http.Response, err error) {
	req, err := http.NewRequest("GET", url, body)

	if err != nil {
		return
	}

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	data,err := parseRes(response.Body, res)
	if err != nil {
		log("err", url, err.Error(), data)
	}
	return
}

func doAndParse(req *http.Request, res interface{}) (*response, error) {
	r := response{}
	response, err := http.DefaultClient.Do(req)
	if response != nil {
		r.StatusCode = response.StatusCode
		r.Status = response.Status
	}
	if err != nil {
		return &r, err
	}
	if response.StatusCode == http.StatusUnauthorized {
		err = newError("doAndParse:StatusUnauthorized")
		return &r,err
	}
	r.Data,err = parseRes(response.Body, res)
	if err != nil {
		log("err", req.URL.String(), response.StatusCode,response.Status,err.Error(), string(r.Data))
	}
	return &r,err
}

func parseRes(body io.Reader, res interface{}) (data []byte, err error) {
	data, _ = ioutil.ReadAll(body)

	err = json.Unmarshal(data, res)
	if err != nil {
		log(string(data))
	}
	return
}

func (d Driver) getRequest(method, url string, body io.Reader) (req *http.Request, err error) {
	if d.err != nil {
		err = d.err
		return
	}
	if d.token == "" {
		err = newError("getRequest:no token")
		return
	}
	req, err = http.NewRequest(method, url, body)
	if err == nil {
		req.Header.Add("authorization", d.token)
	}
	return
}

func makeUrl(uri string) string {
	return apiHost + uri
}

func makeAuthUrl(uri string) string {
	return passportHost + uri
}

func paramBody(param interface{}) io.Reader {
	data, _ := json.Marshal(param)
	return bytes.NewReader(data)
}

