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

func (t twitchServiceHandler) GetDocumentation() serviceHandlerDocumentationList {
	return serviceHandlerDocumentationList{
		{
			ServiceName: "Twitch followers",
			DemoPath:    "/twitch/followers/luziferus",
			Arguments:   []string{"followers", "<user login/id>"},
		},
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
	if len(params) < 2 {
		err = errors.New("No service-command / parameters were given")
		return
	}

	switch params[0] {
	case "followers":
		title, text, color, err = t.handleFollowers(ctx, params[1:])
	case "views":
		title, text, color, err = t.handleViews(ctx, params[1:])
	default:
		err = errors.New("An unknown service command was called")
	}

	return
}

func (t *twitchServiceHandler) handleFollowers(ctx context.Context, params []string) (title, text, color string, err error) {
	followCount, err := t.getUserFollows(params[0])
	if err != nil {
		return "", "", "", errors.Wrap(err, "requesting user follows")
	}

	text = strconv.FormatInt(followCount, 10)
	title = "followers"
	color = "9146FF"

	return
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

	if err := t.doTwitchRequest(http.MethodGet, fmt.Sprintf("https://api.twitch.tv/helix/users?%s=%s", field, params[0]), nil, &respData); err != nil {
		return "", "", "", errors.Wrap(err, "requesting user list")
	}

	if len(respData.Data) != 1 {
		return "", "", "", errors.New("unexpected number of users returned")
	}

	text = strconv.FormatInt(respData.Data[0].ViewCount, 10)
	title = "views"
	color = "9146FF"

	return
}

func (t *twitchServiceHandler) getAccessToken() (string, error) {
	if time.Now().Before(t.accessTokenExpiry) && t.accessToken != "" {
		return t.accessToken, nil
	}

	params := url.Values{}
	params.Set("client_id", configStore.Str(configKeyTwitchClientID))
	params.Set("client_secret", configStore.Str(configKeyTwitchClientSecret))
	params.Set("grant_type", "client_credentials")

	resp, err := http.Post(fmt.Sprintf("https://id.twitch.tv/oauth2/token?%s", params.Encode()), "application/json", nil)
	if err != nil {
		return "", errors.Wrap(err, "request access token")
	}
	defer resp.Body.Close()

	var respData struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", errors.Wrap(err, "reading access token")
	}

	t.accessToken = respData.AccessToken
	t.accessTokenExpiry = time.Now().Add(time.Duration(time.Duration(respData.ExpiresIn)) * time.Second)

	return t.accessToken, nil
}

func (t *twitchServiceHandler) getIDForUser(login string) (string, error) {
	var respData struct {
		Data []struct {
			ID    string `json:"id"`
			Login string `json:"login"`
		} `json:"data"`
	}

	if err := t.doTwitchRequest(http.MethodGet, fmt.Sprintf("https://api.twitch.tv/helix/users?login=%s", login), nil, &respData); err != nil {
		return "", errors.Wrap(err, "requesting user list")
	}

	if len(respData.Data) != 1 {
		return "", errors.New("unexpected number of users returned")
	}

	return respData.Data[0].ID, nil
}

func (t *twitchServiceHandler) getUserFollows(user string) (int64, error) {
	var respData struct {
		Total int64
	}

	if _, err := strconv.ParseInt(user, 10, 64); err != nil {
		if user, err = t.getIDForUser(user); err != nil {
			return 0, errors.Wrap(err, "getting id for user login")
		}
	}

	if err := t.doTwitchRequest(http.MethodGet, fmt.Sprintf("https://api.twitch.tv/helix/users/follows?to_id=%s", user), nil, &respData); err != nil {
		return 0, errors.Wrap(err, "requesting user list")
	}

	return respData.Total, nil
}

func (t *twitchServiceHandler) doTwitchRequest(method, url string, body io.Reader, out interface{}) error {
	at, err := t.getAccessToken()
	if err != nil {
		return errors.Wrap(err, "getting access token")
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	req.Header.Set("Client-Id", configStore.Str(configKeyTwitchClientID))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", at))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "executing request")
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
		return errors.Wrap(err, "reading response")
	}

	return nil
}
