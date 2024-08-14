package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sunoapi/common"
	models "sunoapi/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

const (
	SunoModelChirpV3  = "chirp-v3-0"
	SunoModelChirpV35 = "chirp-v3-5"
)

const RequestIdKey = "X-Any2api-Request-Id"

var Key = "sk-xxx"

func ChatCompletions(c *gin.Context) {
	err := checkChatConfig()
	if err != nil {
		common.ReturnErr(c, err, "suno_config_invalid", http.StatusInternalServerError)
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
		return
	}

	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		Key = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	chatSubmitTemp := common.Templates["chat_stream_submit"]
	chatTickTemp := common.Templates["chat_stream_tick"]
	chatRespTemp := common.Templates["chat_resp"]

	var requestData models.GeneralOpenAIRequest
	err = c.ShouldBindJSON(&requestData)
	if err != nil {
		common.ReturnErr(c, err, "parse_body_failed", http.StatusBadRequest)
		return
	}

	isStream := requestData.Stream
	model := requestData.Model

	switch requestData.Model {
	case "suno-v3":
		model = SunoModelChirpV3
	case "suno-v3.5":
		model = SunoModelChirpV35
	}

	chatID := "chatcmpl-" + c.GetString(RequestIdKey)
	_, params, relayErr := doTools(requestData)
	if relayErr != nil {
		common.ReturnRelayErr(c, relayErr)
		return
	}
	params["mv"] = model

	reqData, _ := json.Marshal(params)
	taskID, err := submitGenSong(reqData)
	if err != nil {
		common.ReturnRelayErr(c, relayErr)
		return
	}

	timeout := time.After(5 * time.Minute)
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	if isStream {
		setSSEHeaders(c)
		c.Writer.WriteHeader(http.StatusOK)
	}

	first := false
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-timeout:
			if isStream {
				common.SendChatData(c.Writer, model, chatID, "timeout")
			} else {
				common.ReturnErr(c, fmt.Errorf("request timeout"), "request_timeout", http.StatusGatewayTimeout)
			}
			return
		case <-tick.C:
			task, relayErr := fetchSong(taskID)
			if relayErr != nil {
				if isStream {
					common.SendChatData(c.Writer, model, chatID, common.GetJsonString(relayErr))
				} else {
					common.ReturnRelayErr(c, relayErr)
				}
				return
			}

			if isStream {
				if !first {
					var songInfo models.SubmitGenSongReq
					_ = json.Unmarshal(reqData, &songInfo)

					if chatSubmitTemp != nil {
						var byteBuf bytes.Buffer
						err = chatSubmitTemp.Execute(&byteBuf, task)
						if err != nil {
							common.SendChatData(c.Writer, model, chatID, common.GetJsonString(common.WrapperErr(err, "template_execute_failed", http.StatusInternalServerError)))
							return
						}
						message := byteBuf.String()
						common.SendChatData(c.Writer, model, chatID, message)
					}
					common.SendChatData(c.Writer, model, chatID, fmt.Sprintf(
						"### 🎵 %v \n\n Tags : %v \n Model : %v \n\n --- \n %v \n > ID \n > %v \n\n 生成中: ",
						songInfo.Title,
						songInfo.Tags,
						songInfo.Mv,
						songInfo.Prompt,
						taskID))
				} else {
					if chatTickTemp != nil {
						var byteBuf bytes.Buffer
						err = chatTickTemp.Execute(&byteBuf, task)
						if err != nil {
							common.SendChatData(c.Writer, model, chatID, common.GetJsonString(common.WrapperErr(err, "template_execute_failed", http.StatusInternalServerError)))
							return
						}
						message := byteBuf.String()
						common.SendChatData(c.Writer, model, chatID, message)
					}
				}
				first = true
			}

			if task.Data.Status.IsDone() {
				var byteBuf bytes.Buffer
				err = chatRespTemp.Execute(&byteBuf, task)
				if err != nil {
					relayErr = common.WrapperErr(err, common.ErrCodeInternalError, http.StatusInternalServerError)
					if isStream {
						common.SendChatData(c.Writer, model, chatID, common.GetJsonString(relayErr))
					} else {
						common.ReturnRelayErr(c, relayErr)
					}
					return
				}

				message := byteBuf.String()
				if isStream {
					common.SendChatData(c.Writer, model, chatID, message)
					common.SendChatDone(c.Writer)
					return
				} else {
					responses := models.ChatCompletionsStreamResponse{
						Id:      chatID,
						Object:  "chat.completion",
						Created: time.Now().Unix(),
						Model:   model,
						Choices: []models.ChatCompletionsStreamResponseChoice{
							{
								Index:        0,
								FinishReason: common.ToPtr("stop"),
								Delta: struct {
									Content string `json:"content"`
									Role    string `json:"role,omitempty"`
								}{
									Content: message,
									Role:    "assistant",
								},
							},
						},
					}
					c.JSON(http.StatusOK, responses)
					return
				}
			}
		}
	}
}

func checkChatConfig() error {
	if common.ChatOpenaiApiBASE == "" {
		return fmt.Errorf("chat_openai_base is empty")
	}
	if common.ChatOpenaiApiKey == "" {
		return fmt.Errorf("chat_openai_key is empty")
	}
	_, ok := common.Templates["chat_resp"]
	if !ok {
		return fmt.Errorf("chat_resp template not found")
	}
	return nil
}

func doTools(requestData models.GeneralOpenAIRequest) (funcName string, res map[string]interface{}, opErr *common.RelayError) {
	requestData.Model = common.ChatOpenaiModel
	requestData.Tools = defaultToolsCalls
	requestData.ToolChoice = "required"
	requestData.Stream = false
	b, err := json.Marshal(requestData)
	if err != nil {
		opErr = common.WrapperErr(err, "request_body_marshal_failed", http.StatusInternalServerError)
		return
	}
	resp, opErr := doOpenAIRequest(bytes.NewBuffer(b), false)
	if opErr != nil {
		return
	}
	defer func() {
		if resp != nil {
			err = resp.Body.Close()
			if err != nil {
				common.LogError(fmt.Sprintf("body close, err: %v", err))
			}
		}
	}()
	body, _ := io.ReadAll(resp.Body)

	// common.LogError(fmt.Sprintf("Response Body: %s", string(body)))

	var openaiResp openai.ChatCompletionResponse
	err = json.Unmarshal(body, &openaiResp)
	if err != nil {
		opErr = common.WrapperErr(err, "tools_body_parse_failed", http.StatusInternalServerError)
		return
	}
	if len(openaiResp.Choices) == 0 || len(openaiResp.Choices[0].Message.ToolCalls) == 0 {
		opErr = common.WrapperErr(err, "no_tools", http.StatusInternalServerError)
		return
	}
	callFunc := openaiResp.Choices[0].Message.ToolCalls[0]
	res = make(map[string]interface{})
	err = json.Unmarshal([]byte(callFunc.Function.Arguments), &res)
	if err != nil {
		opErr = common.WrapperErr(err, "parse_tools_failed", http.StatusBadRequest)
		return
	}
	funcName = callFunc.Function.Name
	return
}

var tagsDescription = `
## tags: The type of song. (Must be in english)
  The following are the example options for each category:
	Style = ['acoustic','aggressive','anthemic','atmospheric','bouncy','chill','dark','dreamy','electronic','emotional','epic','experimental','futuristic','groovy','heartfelt','infectious','melodic','mellow','powerful','psychedelic','romantic','smooth','syncopated','uplifting'];
	Genres = ['afrobeat','anime','ballad','bedroom pop','bluegrass','blues','classical','country','cumbia','dance','dancepop','delta blues','electropop','disco','dream pop','drum and bass','edm','emo','folk','funk','future bass','gospel','grunge','grime','hip hop','house','indie','j-pop','jazz','k-pop','kids music','metal','new jack swing','new wave','opera','pop','punk','raga','rap','reggae','reggaeton','rock','rumba','salsa','samba','sertanejo','soul','synthpop','swing','synthwave','techno','trap','uk garage'];
	Themes = ['a bad breakup','finding love on a rainy day','a cozy rainy day','dancing all night long','dancing with you for the last time','not being able to wait to see you again',"how you're always there for me","when you're not around",'a faded photo on the mantel','a literal banana','wanting to be with you','writing a face-melting guitar solo','the place where we used to go','being trapped in an AI song factory, help!'];
  For example: epic new jack swing`

var promptDesc = `
Lyrics provided in Suno AI V3 optimized format. This format includes a combination structure such as [Intro] [Verse] [Bridge] [Chorus] [Inter] [Inter/solo] [Outro] [Ending], according to the 'Suno AI official instructions', note that about four lines of lyrics per part is the best choice.
The lyrics need to fit the user's description and can be appropriately extended enough to generate a 1 to 3 minute song,thern add a line break between each section.
【 Note 】
Sample lyrics (note the line wrapping format): 
    [Verse]
    City streets, they come alive at night
    Neon lights shining oh so bright (so bright)
    Lost in the rhythm, caught in the beat
    The energy's contagious, can't be discreet (ooh-yeah) \n
    
    [Verse 2]
    Dancin' like there's no tomorrow, we're in the zone
    Fading into the music, we're not alone (alone)
    Feel the passion in every move we make
    We're shaking off the worries, we're wide awake (ooh-yeah) \n
    
    [Chorus]
    Under the neon lights, we come alive (come alive)
    Feel the energy, we're soaring high (soaring high)
    We'll dance until the break of dawn, all through the night (all night)
    Under the neon lights (ooh-ooh-ooh) \n
`

var defaultToolsCalls = []models.Tool{
	{
		Type: string(openai.ToolTypeFunction),
		Function: models.Function{
			Name:        "generate_song_custom",
			Description: "You are sono ai, a songwriting AI.",
			Parameters: models.Parameter{
				Type:     "object",
				Required: []string{"tags", "prompt", "make_instrumental"},
				Properties: map[string]models.Property{
					"make_instrumental": {
						Type:        "boolean",
						Description: "Specifies whether to generate instrumental music tracks? default is false, this property should be set to 'true' only if the user explicitly requests instrumental music.",
					},
					"prompt": {
						Type:        "string",
						Description: promptDesc,
					},
					"title": {
						Type:        "string",
						Description: "The name of the song,",
					},
					"tags": {
						Type:        "string",
						Description: tagsDescription,
					},
					"continue_at": {
						Type:        "string",
						Description: "The time to continue writing in seconds",
					},
					"continue_clip_id": {
						Type:        "string",
						Description: "The id of the song to continue writing",
					},
				},
			},
		},
	},
}

func doOpenAIRequest(requestBody io.Reader, isStream bool) (*http.Response, *common.RelayError) {
	fullRequestURL := fmt.Sprintf("%s/v1/chat/completions", common.ChatOpenaiApiBASE)
	req, err := http.NewRequest(http.MethodPost, fullRequestURL, requestBody)
	if err != nil {
		return nil, common.WrapperErr(err, "resp_body_null", http.StatusInternalServerError)
	}
	req.Header.Set("Authorization", "Bearer "+common.ChatOpenaiApiKey)
	req.Header.Set("Content-Type", "application/json")
	if isStream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}
	resp, err := common.HTTPClient.Do(req)
	if err != nil {
		return nil, common.WrapperErr(err, "do_req_failed", http.StatusInternalServerError)
	}
	if resp == nil {
		return nil, common.WrapperErr(err, "resp_body_null", http.StatusInternalServerError)
	}
	_ = req.Body.Close()
	return resp, nil
}

func setSSEHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}
