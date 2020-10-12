package main

import (
	"context"
	"encoding/json"
	"time"

	"golang.org/x/oauth2/jwt"
)

func idToken(s *ServiceAccount, audience string) (string, error) {
	customClaims := make(map[string]interface{})
	customClaims["target_audience"] = audience

	jwtConfig := jwt.Config{
		Email:         s.ClientEmail,
		PrivateKey:    []byte(s.PrivateKey),
		PrivateKeyID:  s.PrivateKeyID,
		TokenURL:      s.TokenURL,
		UseIDToken:    true,
		PrivateClaims: customClaims,
	}

	ts := jwtConfig.TokenSource(context.Background())

	token, err := ts.Token()
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	Expiry      int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func accessToken(s *ServiceAccount, scopes []string) ([]byte, error) {
	jwtConfig := jwt.Config{
		Email:        s.ClientEmail,
		PrivateKey:   []byte(s.PrivateKey),
		PrivateKeyID: s.PrivateKeyID,
		TokenURL:     s.TokenURL,
		Scopes:       scopes,
		UseIDToken:   false,
	}

	ts := jwtConfig.TokenSource(context.Background())

	token, err := ts.Token()
	if err != nil {
		return nil, err
	}

	accessTokenResponse := AccessTokenResponse{
		AccessToken: token.AccessToken,
		Expiry:      int64(time.Until(token.Expiry).Seconds()),
		TokenType:   token.TokenType,
	}

	data, err := json.Marshal(accessTokenResponse)
	if err != nil {
		return nil, err
	}

	return data, nil
}
