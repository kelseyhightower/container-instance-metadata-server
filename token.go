package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	tokenEndpoint          = "https://oauth2.googleapis.com/token"
	iamcredentialsEndpoint = "https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts"
)

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type AccessTokenRequest struct {
	Scope []string `json:"scope"`
}

type AccessTokenResponse struct {
	AccessToken string    `json:"accessToken"`
	ExpireTime  time.Time `json:"expireTime"`
}

type IDTokenRequest struct {
	Audience     string `json:"audience"`
	IncludeEmail string `json:"includeEmail"`
}

type IDTokenResponse struct {
	Token string `json:"token"`
}

func accessToken() (*Token, error) {
	credentials, err := CredentialsFromFile()
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("grant_type", "refresh_token")
	v.Set("client_id", credentials.ClientID)
	v.Set("client_secret", credentials.ClientSecret)
	v.Set("refresh_token", credentials.RefreshToken)

	request, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error generating user access token: %d", response.StatusCode)
	}

	var t Token
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func generateIdToken(sa, audience string) (string, error) {
	token, err := accessToken()
	if err != nil {
		return "", err
	}

	idTokenRequest := IDTokenRequest{
		Audience:     audience,
		IncludeEmail: "false",
	}

	data, err := json.Marshal(idTokenRequest)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("%s/%s:generateIdToken", iamcredentialsEndpoint, sa)

	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return "", fmt.Errorf("error generating id token: %d", response.StatusCode)
	}

	var idt IDTokenResponse
	if err := json.Unmarshal(data, &idt); err != nil {
		return "", err
	}

	return idt.Token, nil
}

type AccessTokenMetadataResponse struct {
	AccessToken string `json:"access_token"`
	Expiry      int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func generateAccessToken(sa string, scopes []string) ([]byte, error) {
	token, err := accessToken()
	if err != nil {
		return nil, err
	}

	accessTokenRequest := AccessTokenRequest{
		Scope: scopes,
	}

	data, err := json.Marshal(accessTokenRequest)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/%s:generateAccessToken", iamcredentialsEndpoint, sa)

	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("error generating id token: %d", response.StatusCode)
	}

	var a AccessTokenResponse
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}

	md := AccessTokenMetadataResponse{
		AccessToken: a.AccessToken,
		Expiry:      int64(time.Until(a.ExpireTime).Seconds()),
		TokenType:   "Bearer",
	}

	data, err = json.Marshal(md)
	if err != nil {
		return nil, err
	}

	return data, nil
}
