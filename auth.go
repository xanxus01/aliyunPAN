package aliyunDriver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var tmpFile = ""
var bizFile string
var tokenFile string

func TokenGet(code string) {
	param := map[string]string{
		"code":      code,
		"loginType": "normal",
		"deviceId":  "CPH800000000AbfFPI5QSJjO",
	}
	rsp, err := authPost(makeUrl("/token/get"), paramBody(param), nil)
	if err != nil {
		log(err.Error())
		return
	}
	bodyStr, _ := ioutil.ReadAll(rsp.Body)
	log(string(bodyStr))
}

type QrCodeRes struct {
	Content struct {
		Data struct {
			CodeContent string
			CK          string
			ResultCode  int
			T           int
		}
		Success bool
		Status  int
	}
	HasError bool
}

type QrCodeStateRes struct {
	Content struct {
		Data struct {
			QrCodeStatus string
			ResultCode   int
			BizExt       string
		}
		Success bool
		Status  int
	}
	HasError bool
}

func Login() {

}

func QrLogin(qrCodeCallback ...func(codeContent string)) (d *Driver, err error) {
	generateRes, err := Generate()
	if err != nil {
		return
	}
	for _, callback := range qrCodeCallback {
		callback(generateRes.Content.Data.CodeContent)
	}

	ck := generateRes.Content.Data.CK
	t := generateRes.Content.Data.T
	for true {
		qrCodeRes, errQ := QrCodeQuery(ck, t)
		if errQ != nil {
			err = errQ
			break
		}
		qrCodeStatus := qrCodeRes.Content.Data.QrCodeStatus
		log("qrCodeStatus", qrCodeStatus)
		if qrCodeStatus == "NEW" || qrCodeStatus == "SCANED" {
			time.Sleep(2)
			continue
		}
		if qrCodeStatus == "EXPIRED" || qrCodeStatus == "CANCELED" {
			err = newError("QR CODE " + qrCodeStatus)
			break
		}
		if qrCodeStatus == "CONFIRMED" {
			token, errParseBiz := ParseBiz(qrCodeRes.Content.Data.BizExt)
			if errParseBiz != nil {
				return nil, errParseBiz
			}
			d, err = NewDriver(Option{Token: token})
			log(d, err, qrCodeRes)
			break
		}
		err = newError("QR CODE " + qrCodeStatus)
		break
	}
	return
}

func Generate() (res QrCodeRes, err error) {
	urlObj := url.Values{
		"fromSite":    []string{"52"},
		"appName":     []string{"aliyun_drive"},
		"appEntrance": []string{"web"},
		"isMobile":    []string{"false"},
		"lang":        []string{"zh_CN"},
		"returnUrl":   []string{""},
		"bizParams":   []string{""},
	}
	_, err = authGet(makeAuthUrl("/newlogin/qrcode/generate.do?"+urlObj.Encode()), nil, &res)
	if err != nil {
		return
	}

	return
}

func QrCodeQuery(ck string, t int) (res QrCodeStateRes, err error) {
	urlStr := passportHost + "/newlogin/qrcode/query.do?appName=aliyun_drive&fromSite=52&_bx-v=2.2.3"
	params := url.Values{}
	params.Set("ck", ck)
	params.Set("t", strconv.Itoa(t))
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(params.Encode()))

	if err != nil {
		log(err.Error())
		return
	}
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	_, err = doAndParse(req, &res)
	if err != nil {
		log(err.Error())
		return
	}
	//fmt.Printf("%+v", res)
	return
}

func TokenLogin(token string) (string, error) {
	urlStr := authPortHost + "/v2/oauth/token_login"
	params := map[string]string{"token": token}
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	log(string(data))
	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	cookies, err := Authorize()
	if err != nil {
		return "", err
	}
	log(cookies)
	for _, v := range cookies {
		req.AddCookie(v)
	}
	req.Header.Set("content-type", "application/json;charset=UTF-8")
	//req.Header.Set("referer","https://auth.aliyundrive.com/v2/oauth/authorize?client_id=25dzX3vbYqktVxyX&redirect_uri=https%3A%2F%2Fwww.aliyundrive.com%2Fsign%2Fcallback&response_type=code&login_type=custom&state=%7B%22origin%22%3A%22https%3A%2F%2Fwww.aliyundrive.com%22%7D")
	res := make(map[string]string)
	rsp, err := doAndParse(req, &res)

	if err != nil {
		return "", err
	}
	log(res, req.Header.Get("Cookie"), rsp.StatusCode, rsp.StatusCode)
	if res["goto"] == "" {
		err = newError(fmt.Sprintf("goto is empty, url %s ,statusCode %d, status %s, token %s", urlStr, rsp.StatusCode, rsp.Status, token))
		return "", err
	}
	gotoUrl := res["goto"]
	return gotoUrl, err
}

func BizLogin(biz string) (*Driver, error) {
	token, errParseBiz := ParseBiz(biz)
	if errParseBiz != nil {
		return nil, errParseBiz
	}
	d, err := NewDriver(Option{Token: token})
	return d, err
}

func (d *Driver) getToken(code string) error {
	urlStr := apiHost + "/token/get"
	params := map[string]string{
		"code":      code,
		"loginType": "normal",
	}
	req, err := http.NewRequest("POST", urlStr, paramBody(params))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json;charset=UTF-8")

	res := TokenObj{}
	r, err := doAndParse(req, &res)
	if err != nil {
		return err
	}
	log("doAndParse", string(r.Data))
	fmt.Printf("%+v", res)
	data, errJson := json.Marshal(res)
	log("json.Marshal")
	if errJson != nil {
		log(errJson.Error())
	} else {
		log("json.Marshal success")
		Write(string(data), TokenFile())
	}
	d.tokenObj = res
	return err
}

func NewDriverWithTokenObj(obj TokenObj) (*Driver, error) {
	d := &Driver{
		token:    obj.AccessToken,
		tokenObj: obj,
	}
	// Expire
	if obj.ExpireTime.Local().Before(time.Now()) {
		err := d.RefreshToken()
		if err != nil {
			return nil, err
		}
	}
	err := d.sBox()
	if err != nil {
		return nil, err
	}
	return d, nil
}

func Authorize() ([]*http.Cookie, error) {
	urlStr := "https://auth.aliyundrive.com/v2/oauth/authorize?client_id=25dzX3vbYqktVxyX&redirect_uri=https%3A%2F%2Fwww.aliyundrive.com%2Fsign%2Fcallback&response_type=code&login_type=custom&state=%7B%22origin%22%3A%22https%3A%2F%2Fwww.aliyundrive.com%22%7D"
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return response.Cookies(), err
}

func Write(s string, file string) {
	fmt.Println("Write", s)
	f, err := os.OpenFile(file, os.O_CREATE, 0666)
	if err != nil {
		panic(err.Error())
	}
	_, err = f.WriteString(s)
	if err != nil {
		panic(err.Error())
	}
}

func ParseBiz(s string) (string, error) {
	m := make(map[string]map[string]interface{})
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		err = json.Unmarshal(decoded, &m)
	}
	token, ok := (m["pds_login_result"]["accessToken"]).(string)
	if !ok {
		err = newError("token error:" + s)
	}
	return token, err
}

func TmpDir() string {
	if tmpFile != "" {
		return tmpFile
	}
	path, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	tmpFile = path + "\\tmp"
	return tmpFile
}

func init() {
	tmpDir := TmpDir()
	f, err := os.Stat(tmpDir)

	if err != nil {
		panic(err.Error())
	}
	if !f.IsDir() {
		err = os.Mkdir(tmpDir, 0766)
		if err != nil {
			panic(err.Error())
		}
	}
	tokenFile = tmpDir + "\\token.txt"
	bizFile = tmpDir + "\\biz.txt"
}

func TokenFile() string {
	return tokenFile
}

func BizFile() string {
	return bizFile
}

func OpenCodeContent(codeContent string) {
	filePath := "E:\\Documents\\a.png"
	err := qrcode.WriteFile(codeContent, qrcode.Low, 500, filePath)
	if err != nil {
		fmt.Println(err.Error())
	}
	openQrCode(filePath)
}

func openQrCode(path string) {
	cmd := exec.Command("cmd", "/C", path)
	err := cmd.Start()
	if err != nil {
		panic(err.Error())
	}
}
