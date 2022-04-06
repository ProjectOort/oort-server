package account

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ProjectOort/oort-server/conf"
)

type giteeHelper struct {
	httpClient *http.Client

	clientID     string
	clientSecret string
	redirectURI  string
}

func newGiteeHelper(cfg *conf.Account) *giteeHelper {
	return &giteeHelper{
		httpClient:   &http.Client{},
		clientID:     cfg.GiteeClientID,
		clientSecret: cfg.GiteeClientSecret,
		redirectURI:  cfg.GiteeRedirectURI,
	}
}

const _OAuthURL = "https://gitee.com/oauth/token"

type oauthResult struct {
	AccessToken  string `json:"access_token"`
	CreatedAt    int    `json:"created_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func (x *giteeHelper) OAuth(code string) (*oauthResult, error) {
	var params = struct {
		GrantType    string `json:"grant_type"`
		Code         string `json:"code"`
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		RedirectURI  string `json:"redirect_uri"`
	}{
		GrantType:    "authorization_code",
		Code:         code,
		ClientID:     x.clientID,
		ClientSecret: x.clientSecret,
		RedirectURI:  x.redirectURI,
	}

	paramsBytes, _ := json.Marshal(params)

	// send request
	response, err := x.httpClient.Post(_OAuthURL, "application/json", bytes.NewBuffer(paramsBytes))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("invalid grant")
	}

	// read response
	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// combine result
	var result oauthResult
	err = json.Unmarshal(responseBytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

const _UserInfoURL = "https://gitee.com/api/v5/user"

type userInfoResult struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	URL       string `json:"url"`
	HTMLURL   string `json:"html_url"`
	Remark    string `json:"remark"`
}

func (x *giteeHelper) UserInfo(accessToken string) (*userInfoResult, error) {
	response, err := x.httpClient.Get(fmt.Sprintf("%s?access_token=%s", _UserInfoURL, accessToken))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.New("invalid status")
	}

	responseBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var result userInfoResult
	err = json.Unmarshal(responseBytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
