package vkapi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"

	"github.com/json-iterator/go"
	"go.uber.org/ratelimit"
)

var json = jsoniter.ConfigFastest

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
	ApiDomain   string
}

// Create a new api client with given access_token
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
		ApiDomain:   apiDomain,
	}
}

// Executes GET request to API method with given query parameters. Argument param can be nil if there is no parameters.
//
// There is 2 types of errors:
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

// Executes GET request to API method with given query parameters. Argument param can be nil if there is no parameters.
//
// There is 2 types of errors:
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
	return uploadRequest(serverUrl, contentType, b)
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
	return uploadRequest(serverUrl, contentType, b)
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
	return uploadRequest(serverUrl, contentType, b)
}

func makeMultipartData(fieldCreator func(writer *multipart.Writer) (io.Writer, error), writer func(io.Writer) error) (*bytes.Buffer, string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := fieldCreator(w)
	if err != nil {
		return nil, "", err
	}
	err = writer(fw)
	if err != nil {
		return nil, "", err
	}
	w.Close()
	return &b, w.FormDataContentType(), nil
}

func uploadRequest(serverUrl, contentType string, requestBody *bytes.Buffer) ([]byte, error) {
	request, err := http.NewRequest("POST", serverUrl, requestBody)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}
