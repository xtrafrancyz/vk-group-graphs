package vkapi

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"os"

	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigFastest

type Api struct {
	accessToken string
	ApiDomain   string
}

func Create(token string) *Api {
	apiDomain := os.Getenv("VK_API_DOMAIN")
	if apiDomain == "" {
		apiDomain = "api.vk.com"
	}
	return &Api{
		accessToken: token,
		ApiDomain:   apiDomain,
	}
}

func (api *Api) Request(method string, params map[string]string) (map[string]interface{}, error) {
	requestUrl, err := url.Parse("https://" + api.ApiDomain + "/method/" + method)
	if err != nil {
		return nil, err
	}
	requestQuery := requestUrl.Query()
	for key, value := range params {
		requestQuery.Set(key, value)
	}
	requestQuery.Set("access_token", api.accessToken)

	requestUrl.RawQuery = requestQuery.Encode()

	response, err := http.Get(requestUrl.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(content, &parsed); err == nil {
		if errorObj, ok := parsed["error"]; ok {
			errorCasted := errorObj.(map[string]interface{})
			return parsed, &ApiError{
				Message: errorCasted["error_msg"].(string),
				Method:  method,
				Params:  &params,
				Code:    int(errorCasted["error_code"].(float64)),
			}
		}
		return parsed, nil
	}
	return nil, err
}

type ApiError struct {
	Message string
	Method  string
	Params  *map[string]string
	Code    int
}

func (e *ApiError) Error() string {
	return e.Message
}
