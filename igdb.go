package igdb

import (
	"encoding/json"
	"github.com/Henry-Sarabia/apicalypse"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
)

// igdbURL is the base URL for the IGDB API.
const igdbURL string = "https://api-v3.igdb.com/"

// Errors returned when creating URLs for API calls.
var (
	// ErrNegativeID occurs when a negative ID is used as an argument in an API call.
	ErrNegativeID = errors.New("igdb.Client: negative ID")
	// ErrNoResults occurs when the IGDB returns no results
	ErrNoResults = errors.New("igdb.Client: no results")
)

// service is the underlying struct that handles
// all API calls for different IGDB endpoints.
type service struct {
	client *Client
}

// Client wraps an HTTP Client used to communicate with the IGDB,
// the root URL of the IGDB, and the user's IGDB API key.
// Client also initializes all the separate services to communicate
// with each individual IGDB API endpoint.
type Client struct {
	http    *http.Client
	rootURL string
	key     string

	common service

	// Services
	Games *GameService
}

// NewClient returns a new Client configured to communicate with the IGDB.
// The provided apiKey will be used to make requests on your behalf. The
// provided HTTP Client will be the client making requests to the IGDB.
// If no HTTP Client is provided, a default HTTP client is used instead.
//
// If you need an IGDB API key, please visit: https://api.igdb.com/signup
func NewClient(apiKey string, custom *http.Client) *Client {
	if custom == nil {
		custom = http.DefaultClient
	}
	c := &Client{}
	c.http = custom
	c.key = apiKey
	c.rootURL = igdbURL

	c.common.client = c
	c.Games = (*GameService)(&c.common)

	return c
}

// Request configures a new request for the provided URL and
// adds the necesarry headers to communicate with the IGDB.
func (c *Client) request(end endpoint, opts ...FuncOption) (*http.Request, error) {
	unwrapped, err := unwrapOptions(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create request with invalid options")
	}

	req, err := apicalypse.NewRequest("GET", c.rootURL+string(end), unwrapped...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot make request for '%s' endpoint", end)
	}

	req.Header.Add("user-key", c.key)
	req.Header.Add("Accept", "application/json")

	return req, nil
}

// Send sends the provided request and stores the response in the value pointed to by result.
// The response will be checked and return any errors.
func (c *Client) send(req *http.Request, result interface{}) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrap(err, "http client cannot send request")
	}
	defer resp.Body.Close()

	if err = checkResponse(resp); err != nil {
		return err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read response body")
	}

	if err = checkResults(b); err != nil {
		return err
	}

	err = json.Unmarshal(b, &result)

	return err
}

// Get sends a GET request to the provided endpoint with the provided options and
// stores the results in the value pointed to by result.
func (c *Client) get(end endpoint, result interface{}, opts ...FuncOption) error {
	req, err := c.request(end, opts...)
	if err != nil {
		return err
	}

	err = c.send(req, result)
	if err != nil {
		return errors.Wrap(err, "cannot make GET request")
	}

	return nil
}
