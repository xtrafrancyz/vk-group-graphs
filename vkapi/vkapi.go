package vkapi

import (
	"net/url"
	"net/http"
	"io/ioutil"
	"os"
	"fmt"

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

func (api *Api) Request(method string, params map[string]string) ([]byte, error) {
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

	errorVal := json.Get(content, "error")
	if errorVal.ValueType() == jsoniter.ObjectValue {
		return content, &ApiError{
			Message: errorVal.Get("error_msg").ToString(),
			Method:  method,
			Params:  params,
			Code:    errorVal.Get("error_code").ToInt(),
		}
	}

	return content, nil
}

func (api *Api) RequestJson(method string, params map[string]string) (map[string]interface{}, error) {
	content, err := api.Request(method, params)
	if content != nil {
		var parsed map[string]interface{}
		err2 := json.Unmarshal(content, &parsed)
		if err2 != nil {
			return nil, err2
		}
		return parsed, err
	}
	return nil, err
}

type ApiError struct {
	Message string
	Method  string
	Params  map[string]string
	Code    int
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("VkApiError. Method: %s, ErrorCode: %d, %s", e.Method, e.Code, e.Message)
}
