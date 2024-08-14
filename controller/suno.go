package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sunoapi/common"
	models "sunoapi/model"

	"github.com/gin-gonic/gin"
)

var CommonHeaders = map[string]string{
	"Content-Type": "application/json",
	"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Accept":       "*/*",
}

func MakeRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if body != nil {
			err = req.Body.Close()
			if err != nil {
				common.LogError(fmt.Sprintf("body close err:%v", err.Error()))
			}
		}
	}()
	for k, v := range CommonHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+Key)
	// common.LogSuccess(req.Header["Authorization"][0])

	resp, err := common.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func submitGenSong(reqBody []byte) (taskID string, err error) {
	var params models.SubmitGenSongReq
	err = json.Unmarshal(reqBody, &params)
	if err != nil {
		return "", err
	}
	common.LogSuccess(fmt.Sprintf("generateSong get requestBody\n:%v\n", string(reqBody)))

	reqData := make(map[string]interface{})
	reqData["make_instrumental"] = params.MakeInstrumental
	if params.Mv != "" {
		reqData["mv"] = params.Mv
	} else {
		reqData["mv"] = "chirp-v3-0"
	}
	if params.GptDescriptionPrompt != "" {
		reqData["gpt_description_prompt"] = params.GptDescriptionPrompt
		reqData["prompt"] = ""
	} else {
		reqData["prompt"] = params.Prompt
		reqData["title"] = params.Title
		reqData["tags"] = params.Tags
		reqData["continue_at"] = params.ContinueAt
		reqData["continue_clip_id"] = params.ContinueClipId
	}
	if params.ContinueClipId != nil && *params.ContinueClipId != "" { // 续写
		if params.TaskID == "" {
			return "", fmt.Errorf("task_id is empty")
		}
	}

	// Logging request body
	requestBody, _ := json.Marshal(reqData)
	// common.LogSuccess(fmt.Sprintf("Request Body: %s", string(requestBody)))
	// Send request
	resp, err := MakeRequest("POST", fmt.Sprintf("%s/suno/submit/music", common.BaseUrl), bytes.NewReader(requestBody), nil)
	if err != nil {
		fmt.Printf("Request failed: %s\n", err.Error())
		return
	}
	defer resp.Body.Close()
	// Read response body
	var responseFeed models.ResponseMusic
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err.Error())
		return
	}

	// Log the response body
	// common.LogSuccess(fmt.Sprintf("-----%s------", Key))
	common.LogSuccess(fmt.Sprintf("Response Body: %s", string(bodyBytes)))

	// Unmarshal response body
	err = json.Unmarshal(bodyBytes, &responseFeed)
	if err != nil {
		return "", err
	}
	return responseFeed.Data, nil
}
func fetchSong(taskID string) (data models.ResponseFeed, relayErr *common.RelayError) {
	url := fmt.Sprintf("%s/suno/fetch/%v", common.BaseUrl, taskID)
	resp, err := MakeRequest("GET", url, nil, nil)
	if err != nil {
		return data, common.WrapperErr(fmt.Errorf("fetchSong:Request failed:%s", err.Error()), common.ErrCodeInvalidRequest, http.StatusBadRequest)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, common.WrapperErr(fmt.Errorf("fetchSong:Failed to read response body:%s", err.Error()), common.ErrCodeInvalidRequest, http.StatusBadRequest)
	}
	// common.LogSuccess(string(bodyBytes))
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return data, common.WrapperErr(fmt.Errorf("task not exist"), common.ErrCodeInvalidRequest, http.StatusBadRequest)
	}
	return data, nil
}

func FetchTask(c *gin.Context) {
	taskID := c.Param("task_id")
	chatRespTemp := common.Templates["chat_resp"]
	task, err := fetchSong(taskID)
	if err != nil {
		c.JSON(400, gin.H{
			"code":  err.Code,
			"error": err.Err.Error(),
		})
	}
	var byteBuf bytes.Buffer
	excuteErr := chatRespTemp.Execute(&byteBuf, task)
	if excuteErr != nil {
		c.JSON(400, gin.H{
			"code":  400,
			"error": excuteErr.Error(),
		})
	}

}
