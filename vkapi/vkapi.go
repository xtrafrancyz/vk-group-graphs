package vkapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"time"

	"go.uber.org/ratelimit"
)

type errorResponse struct {
	Error struct {
		ErrorMessage string `json:"error_message"`
		ErrorCode    int    `json:"error_code"`
	} `json:"error"`
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

type Api struct {
	accessToken string
	version     string
	rateLimiter ratelimit.Limiter
	httpClient  *http.Client
	ApiDomain   string
}

// CreateWithToken create a new api client with given access_token
// You can use custom api domain with env variable VK_API_DOMAIN
func CreateWithToken(token, version string) *Api {
	apiDomain := os.Getenv("VK_API_DOMAIN")
	if apiDomain == "" {
		apiDomain = "api.vk.com"
	}
	return &Api{
		accessToken: token,
		version:     version,
		rateLimiter: ratelimit.New(3, ratelimit.WithoutSlack),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		ApiDomain: apiDomain,
	}
}

// Request executes GET request to API method with given query parameters. Argument param can be nil if there is no parameters.
//
// There are 2 types of errors:
// 1. Any IO error will occur.
// 2. There is an error in response from VK API. You can cast it to ApiError and see what happens. Response from the API
// will also be included in the return values.
func (api *Api) Request(method string, params map[string]string) ([]byte, error) {
	api.rateLimiter.Take()

	requestUrl, err := url.Parse("https://" + api.ApiDomain + "/method/" + method)
	if err != nil {
		return nil, err
	}
	requestQuery := requestUrl.Query()
	if params != nil {
		for key, value := range params {
			requestQuery.Set(key, value)
		}
	}
	requestQuery.Set("access_token", api.accessToken)
	requestQuery.Set("v", api.version)
	requestUrl.RawQuery = requestQuery.Encode()

	response, err := api.httpClient.Get(requestUrl.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var possibleError errorResponse
	if err = json.Unmarshal(content, &possibleError); possibleError.Error.ErrorCode != 0 {
		return content, &ApiError{
			Message: possibleError.Error.ErrorMessage,
			Method:  method,
			Params:  params,
			Code:    possibleError.Error.ErrorCode,
		}
	}

	return content, nil
}

// RequestJson executes GET request to API method with given query parameters. Argument param can be nil if there is no parameters.
//
// There are 2 types of errors:
// 1. Any IO error will occur.
// 2. There is an error in response from VK API. You can cast it to ApiError and see what happens. Response from the API
// will also be included in the return values.
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

// RequestJsonStruct executes GET request to API method with given query parameters. Argument param can be nil if there is no parameters.
//
// There are 2 types of errors:
// 1. Any IO error will occur.
// 2. There is an error in response from VK API. You can cast it to ApiError and see what happens.
func (api *Api) RequestJsonStruct(method string, params map[string]string, val any) error {
	content, err := api.Request(method, params)
	if content != nil {
		err2 := json.Unmarshal(content, val)
		if err2 != nil {
			return err2
		}
		return err
	}
	return err
}

func (api *Api) Upload(serverUrl, field string, file []byte) ([]byte, error) {
	b, contentType, err := makeMultipartData(
		func(writer *multipart.Writer) (io.Writer, error) {
			return writer.CreateFormField(field)
		},
		func(writer io.Writer) error {
			_, err := writer.Write(file)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return api.uploadRequest(serverUrl, contentType, b)
}

func (api *Api) UploadStream(serverUrl, field string, reader io.Reader) ([]byte, error) {
	b, contentType, err := makeMultipartData(
		func(writer *multipart.Writer) (io.Writer, error) {
			return writer.CreateFormField(field)
		},
		func(writer io.Writer) error {
			_, err := io.Copy(writer, reader)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return api.uploadRequest(serverUrl, contentType, b)
}

func (api *Api) UploadFile(serverUrl, field string, file *os.File) ([]byte, error) {
	b, contentType, err := makeMultipartData(
		func(writer *multipart.Writer) (io.Writer, error) {
			return writer.CreateFormFile(field, file.Name())
		},
		func(writer io.Writer) error {
			_, err := io.Copy(writer, file)
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return api.uploadRequest(serverUrl, contentType, b)
}

func (api *Api) uploadRequest(serverUrl, contentType string, requestBody *bytes.Buffer) ([]byte, error) {
	request, err := http.NewRequest("POST", serverUrl, requestBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	response, err := api.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func makeMultipartData(fieldCreator func(writer *multipart.Writer) (io.Writer, error), writer func(io.Writer) error) (*bytes.Buffer, string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := fieldCreator(w)
	if err != nil {
		return nil, "", err
	}
	defer w.Close()
	err = writer(fw)
	if err != nil {
		return nil, "", err
	}
	return &b, w.FormDataContentType(), nil
}
