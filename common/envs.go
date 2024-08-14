package common

import (
	"net/http"
	"time"
)

var HTTPClient *http.Client
var StartTime = time.Now().Unix() // unit: second
var Port = GetOrDefaultString("PORT", "8000")

var SecretToken = GetOrDefaultString("SECRET_TOKEN", "")

var ChatTemplateDir = GetOrDefaultString("CHAT_TEMPLATE_DIR", "./template")
var BaseUrl = GetOrDefaultString("BASE_URL", "https://studio-api.suno.ai")
var ChatOpenaiModel = GetOrDefaultString("CHAT_OPENAI_MODEL", "gpt-4o")
var ChatOpenaiApiBASE = GetOrDefaultString("CHAT_OPENAI_BASE", "https://api.openai.com")
var ChatOpenaiApiKey = GetOrDefaultString("CHAT_OPENAI_KEY", "")
