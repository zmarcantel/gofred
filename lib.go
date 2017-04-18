package gofred

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// Base API endpoint URL all requests go through
	API_URL = "https://api.stlouisfed.org/fred"
)

//==============================================================================
// client config
//==============================================================================

// Enum type that allows selecting the response format.
type ResponseFormat uint8

const (
	// Get responses in JSON format
	JSON ResponseFormat = iota
	// Get responses in XML format
	XML ResponseFormat = iota + 1
)

// Get the string representation of the response format as it should be used in a URL param.
func (f ResponseFormat) String() string {
	switch f {
	case JSON:
		return "json"
	case XML:
		return "xml"
	default:
		return "ERROR" // TODO
	}
}

//==============================================================================
// requests
//==============================================================================

// Interface to a generic request to the Fred API.
type Request interface {
	ToParams() url.Values
}

// Minimal shared request objects.
//
// A `BaseRequest` object is held in the client and copied for every request
// generated through the library. Once copied, the request-specific parameters
// are filled in.
type BaseRequest struct {
	fmt     ResponseFormat
	api_key string
}

// Satisfies the `Request` interface, generating a `url.Values` set with the API key and format.
func (r BaseRequest) ToParams() url.Values {
	v := url.Values{}
	v.Set("api_key", r.api_key)
	v.Set("file_type", r.fmt.String())
	return v
}

//==============================================================================
// responses
//==============================================================================

// Generic error response type.
//
// If a non-success return code is returned, this type is expected to be parseable.
type BaseError struct {
	Message string `json:"error_message" xml:"error_message"`
	Code    uint32 `json:"error_code" xml:"error_code"`
}

//==============================================================================
// client
//==============================================================================

// Main interface to the API.
//
// Requires specifying the API key and response format for all future requests
// through this client.
type Client struct {
	base_req BaseRequest
	base_url url.URL
}

// Create a new client with the given API key and response format.
func NewClient(key string, format ResponseFormat) (Client, error) {
	api_url, err := url.Parse(API_URL)
	if err != nil {
		return Client{}, err
	}

	return Client{
		base_req: BaseRequest{
			fmt:     format,
			api_key: key,
		},
		base_url: *api_url,
	}, nil
}

// Unmarshals the byte slice into the target interface based on the internal
// response format given when the client was created.
func (c Client) unmarshal_body(body []byte, into interface{}) error {
	switch c.base_req.fmt {
	case JSON:
		err := json.Unmarshal(body, into)
		if err != nil {
			return fmt.Errorf("failed to parse json response: %v", err)
		}
	case XML:
		err := xml.Unmarshal(body, into)
		if err != nil {
			return fmt.Errorf("failed to parse xml response: %v", err)
		}
	default:
		return fmt.Errorf("unknown request/response type: %v", c.base_req.fmt)
	}

	return nil
}

// Parses the byte slice as a `BaseError` depending on the response format.
func (c Client) get_error(body []byte) (BaseError, error) {
	var result BaseError
	switch c.base_req.fmt {
	case JSON:
		err := json.Unmarshal(body, &result)
		if err != nil {
			return BaseError{}, fmt.Errorf("failed to parse json error response: %v", err)
		}
	case XML:
		err := xml.Unmarshal(body, &result)
		if err != nil {
			return BaseError{}, fmt.Errorf("failed to parse xml error response: %v", err)
		}
	default:
		return BaseError{}, fmt.Errorf("unknown request/response type: %v", c.base_req.fmt)
	}

	return result, nil
}

// Wrapper around `http.Get()` which checks status codes and proxies back either a
// valid response or a parsed/generated error.
func (c Client) get(desc, req_url string) ([]byte, error) {
	res, err := http.Get(req_url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// catch early errors
	switch res.StatusCode {
	case 200:
		// do nothing
	case 404:
		return nil, fmt.Errorf("could not find %s: %d", desc, res.StatusCode)
	default:
		req_err, err := c.get_error(body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s error response: %v", desc, err)
		}
		return nil, fmt.Errorf("could not get %s (%d): %v", desc, req_err.Code, req_err.Message)
	}

	return body, nil
}
