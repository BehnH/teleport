package dep

import (
	"bytes"
	"encoding/json"
	"github.com/gomodule/oauth1/oauth"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	defaultBaseUrl        = "https://mdmenrollment.apple.com/"
	contentType           = "application/json;charset=UTF8"
	XServerProtocolHeader = "X-Server-Protocol-Version"
	ProtocolVersionNumber = "3"
	TeleportVersion = "v0.0.1-ALPHA"
)

type Client struct {
	ConsumerKey      string // Provided by Apple
    ConsumerSecret   string // Provided by Apple
    AccessToken      string // Provided by Apple
    AccessSecret     string // Provided by Apple
	AuthSessionToken string // Generated using CreateSession() function, with above values
	SessionExpires   time.Time

	UserAgent string
	Client    HTTPClient
	BaseUrl   *url.URL

	SessionMutex sync.Mutex
}

type Option func(*Client)

type OAuthParams struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
	AccessToken    string `json:"access_token"`
	AccessSecret   string `json:"access_secret"`
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func WithServerUrl(baseUrl *url.URL) Option {
    return func (c *Client) {
        c.BaseUrl = baseUrl
    }
}

func WithHttpClient(client HTTPClient) Option {
    return func (c *Client) {
        c.Client = client
    }
}

func CreateClient(params OAuthParams, opts ...Option) *Client {
	baseUrl, _ := url.Parse(defaultBaseUrl)
	client := Client{
		ConsumerKey:    params.ConsumerKey,
		ConsumerSecret: params.ConsumerSecret,
		AccessToken:    params.AccessToken,
		AccessSecret:   params.AccessSecret,

		UserAgent: path.Join("teleport_mdm", TeleportVersion),
		BaseUrl:   baseUrl,
	}

	for _, optnFn := range opts {
		optnFn(&client)
	}

	return &client
}

func (client *Client) GetSession() error {
	// Ensure the operation is thread-safe JIC it's used across multiple routines
	client.SessionMutex.Lock()
	defer client.SessionMutex.Unlock()

	// If there is no AuthSessionToken, generate one
	if client.AuthSessionToken == "" {
		if err := client.CreateSession(); err != nil {
			return errors.Wrap(err, "Failed to create new session with DEP server")
		}
	}

	// If the token has expired, refresh it
	if time.Now().After(client.SessionExpires) {
		if err := client.CreateSession(); err != nil {
			return errors.Wrap(err, "Failed to create new auth session with DEP server")
		}
	}

	return nil
}

func (client *Client) CreateSession() error {
	var authSessionToken struct {
		AuthSessionToken string `json:"auth_session_token"`
	}
	// Create the consumer credentials config
	consumerCreds := oauth.Credentials{
		Token: client.ConsumerKey,
		Secret: client.ConsumerSecret,
	}
	// Create the access token
	accessToken := &oauth.Credentials{
		Token: client.AccessToken,
		Secret: client.AccessSecret,
	}

	// Create an empty map of form values
	form := url.Values{}

	// Ensure the URL path is valid
	path, err := url.Parse("/session")
	if err != nil {
		return err
	}

	// Build the session URL
	sessionUrl := client.BaseUrl.ResolveReference(path)

	// Create the actual consumer client
	consumerClient := oauth.Client{
		SignatureMethod: oauth.HMACSHA1,
		TokenRequestURI: sessionUrl.String(),
		Credentials:     consumerCreds,
	}

	// Create the GET request
	req, err := http.NewRequest(http.MethodGet, consumerClient.TokenRequestURI, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	if err := consumerClient.SetAuthorizationHeader(
		req.Header,
		accessToken,
		http.MethodGet,
		req.URL,
		form,
	); err != nil {
		return err
	}

	req.Header.Add("User-Agent", client.UserAgent)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add(XServerProtocolHeader, ProtocolVersionNumber)

	// Get the Auth header
	res, err := client.Client.Do(req)
	if err != nil {
		return err
	}
	// Defer closing the connection
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("DEP server did not provide OK. Returned: %v", res.Status)
	}

	if err := json.NewDecoder(res.Body).Decode(&authSessionToken); err != nil {
		return errors.Wrap(err, "decoding AuthSessionToken from DEP response")
	}

	// Set the AuthSessionToken
	client.AuthSessionToken = authSessionToken.AuthSessionToken
	// Set the session expiry time
	client.SessionExpires = time.Now().Add(time.Minute * 3)

	return nil
}

func (client *Client) CreateRequest(method, urlString string, body interface{}) (*http.Request, error) {
	path, err := url.Parse(urlString)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse DEP request path: %s", urlString)
	}

	url := client.BaseUrl.ResolveReference(path)

	var buffer bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buffer).Encode(body); err != nil {
			return nil, errors.Wrapf(err, "Failed to encode DEP request body")
		}
	}

	req, err := http.NewRequest(method, url.String(), &buffer)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create %s request to DEP server %s", method, url.String())
	}

	req.Header.Add("User-Agent", client.UserAgent)
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Accept", contentType)
	req.Header.Add(XServerProtocolHeader, ProtocolVersionNumber)
	return req, nil
}

func (client *Client) do(req *http.Request, to interface{}) error {
	if err := client.GetSession(); err != nil {
		return errors.Wrapf(err, "Failed to get session for request to %s", client.BaseUrl.String())
	}

	req.Header.Add("X-ADM-Auth-Session", client.AuthSessionToken)

	res, err := client.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed to perform DEP request")
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return errors.Errorf("Unexpected response received from DEP server. Response=%d | DEP API Error: %s", res.StatusCode, string(body))
	}

    err = json.NewDecoder(res.Body).Decode(to)
    return errors.Wrap(err, "Failed to decode DEP response body")
}
