package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	// #configStore twitch.client_id - string - ID of a Twitch application
	configKeyTwitchClientID = "twitch.client_id"
	// #configStore twitch.client_secret - string - Secret of the Twitch application identified by twitch.client_id
	configKeyTwitchClientSecret = "twitch.client_secret"
)

func init() {
	registerServiceHandler("twitch", &twitchServiceHandler{})
}

type twitchServiceHandler struct {
	accessToken       string
	accessTokenExpiry time.Time
}

func (twitchServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "Twitch views",
			DemoPath:    "/twitch/views/luziferus",
			Arguments:   []string{"views", "<user login/id>"},
		},
	}
}

func (twitchServiceHandler) IsEnabled() bool {
	return configStore[configKeyTwitchClientID] != nil && configStore[configKeyTwitchClientSecret] != nil
}

func (t *twitchServiceHandler) Handle(ctx context.Context, params []string) (title, text, color string, err error) {
	if len(params) < 2 { //nolint:gomnd
		err = errors.New("No service-command / parameters were given")
		return title, text, color, err
	}

	switch params[0] {
	case "views":
		title, text, color, err = t.handleViews(ctx, params[1:])
	default:
		err = errors.New("An unknown service command was called")
	}

	return title, text, color, err
}

func (t *twitchServiceHandler) handleViews(ctx context.Context, params []string) (title, text, color string, err error) {
	var respData struct {
		Data []struct {
			Login     string `json:"login"`
			ViewCount int64  `json:"view_count"`
		} `json:"data"`
	}

	field := "login"
	if _, err := strconv.ParseInt(params[0], 10, 64); err == nil {
		field = "id"
	}

	if err := t.doTwitchRequest(ctx, http.MethodGet, fmt.Sprintf("https://api.twitch.tv/helix/users?%s=%s", field, params[0]), nil, &respData); err != nil {
		return "", "", "", errors.Wrap(err, "requesting user list")
	}

	if len(respData.Data) != 1 {
		return "", "", "", errors.New("unexpected number of users returned")
	}

	text = strconv.FormatInt(respData.Data[0].ViewCount, 10)
	title = "views"
	color = "9146FF"

	return title, text, color, err
}

func (t *twitchServiceHandler) getAccessToken(ctx context.Context) (string, error) {
	if time.Now().Before(t.accessTokenExpiry) && t.accessToken != "" {
		return t.accessToken, nil
	}

	params := url.Values{}
	params.Set("client_id", configStore.Str(configKeyTwitchClientID))
	params.Set("client_secret", configStore.Str(configKeyTwitchClientSecret))
	params.Set("grant_type", "client_credentials")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()), nil)
	if err != nil {
		return "", errors.Wrap(err, "creating access token request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "executing access token request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	var respData struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", errors.Wrap(err, "reading access token")
	}

	t.accessToken = respData.AccessToken
	t.accessTokenExpiry = time.Now().Add(time.Duration(respData.ExpiresIn) * time.Second)

	return t.accessToken, nil
}

func (t *twitchServiceHandler) doTwitchRequest(ctx context.Context, method, reqURL string, body io.Reader, out any) error {
	at, err := t.getAccessToken(ctx)
	if err != nil {
		return errors.Wrap(err, "getting access token")
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	req.Header.Set("Client-Id", configStore.Str(configKeyTwitchClientID))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", at))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "executing request")
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing response body (leaked fd)")
		}
	}()

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return errors.Wrap(err, "reading response")
	}

	return nil
}
