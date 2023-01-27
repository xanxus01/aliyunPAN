package aliyunDriver

import (
	"errors"
	"net/url"
	"time"
)

const apiHost = "https://api.aliyundrive.com"
const passportHost = "https://passport.aliyundrive.com"
const authPortHost = "https://auth.aliyundrive.com"

type Driver struct {
	token        string
	err          error
	DriveId      string
	refreshToken string
	tokenObj     TokenObj
}

type TokenObj struct {
	DefaultSboxDriveId string `json:"default_sbox_drive_id"`
	Role               string `json:"role"`
	UserName           string `json:"user_name"`
	NeedLink           bool   `json:"need_link"`
	ExpireTime         time.Time `json:"expire_time"`
	PinSetup           bool   `json:"pin_setup"`
	NeedRpVerify       bool   `json:"need_rp_verify"`
	Avatar             string `json:"avatar"`
	UserData           struct {
		DingDingRobotUrl string `json:"ding_ding_robot_url"`
		EncourageDesc    string `json:"encourage_desc"`
		FeedBackSwitch   bool   `json:"feed_back_switch"`
		FollowingDesc    string `json:"following_desc"`
	}
	TokenType      string   `json:"token_type"`
	AccessToken    string   `json:"access_token"`
	DefaultDriveId string   `json:"default_drive_id"`
	DomainId       string   `json:"domain_id"`
	RefreshToken   string   `json:"refresh_token"`
	IsFirstLogin   bool     `json:"is_first_login"`
	UserId         string   `json:"user_id"`
	NickName       string   `json:"nick_name"`
	ExistLink      []string `json:"exist_link"`
	State          string   `json:"state"`
	ExpiresIn      int      `json:"expires_in"`
	Status         string   `json:"status"`
}

type Option struct {
	Token string
}

type Result struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Res interface {
	Success() bool
}

func (r Result) Success() bool {
	return r.Code == ""
}

func UseTokenLogin(token string)  {

}

func NewDriver(option Option) (d *Driver, err error) {
	if option.Token == "" {
		err = newError("need token")
	} else {
		gotoUrl := ""
		gotoUrl, err = TokenLogin(option.Token)
		if err != nil {
			return
		}
		utlObj, err1 := url.Parse(gotoUrl)
		if err1 != nil {
			log(err1)
			err = err1
			return
		}
		val := utlObj.Query()
		d = &Driver{token: option.Token}
		err = d.getToken(val.Get("code"))
		if err != nil {
			return nil, err
		}
		err = d.sBox()

	}
	return
}

func (d Driver) sBox() (err error) {
	res := struct {
		DriveId          string `json:"drive_id"`
		InsuranceEnabled bool `json:"insurance_enabled"`
		Locked           bool `json:"locked"`
		PinSetup         bool `json:"pin_setup"`
		RecommendVip     string `json:"recommend_vip"`
		SBoxRealUsedSize int `json:"s_box_real_used_size"`
		SBoxTotalSize    int `json:"s_box_total_size"`
		SBoxUsedSize     int `json:"s_box_used_size"`
	}{}
	_, err = d.post(makeUrl("/v2/sbox/get"), nil, &res)
	if err != nil {
		d.DriveId = res.DriveId
	} else {
		d.err = nil
	}
	return err
}

func (d Driver)GetTokenObj() TokenObj {
	return d.tokenObj
}

func (d *Driver)RefreshToken() error {
	m := map[string]string{
		"refresh_token":d.tokenObj.RefreshToken,
	}
	res := TokenObj{}
	r,err := d.postJson(makeUrl("/token/refresh"), paramBody(m), &res)
	if err != nil {
		return err
	}

	if res.AccessToken == "" {
		return errors.New("error " + string(r.Data))
	}

	Write(string(r.Data), TokenFile())
	d.tokenObj = res
	d.token = res.AccessToken
	return nil
}
