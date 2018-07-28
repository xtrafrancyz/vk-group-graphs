package main

import (
	"bytes"
	"log"

	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

var json = jsoniter.ConfigFastest
var methodPost = []byte("POST")

type WebServer struct {
	confirmationKey string
	secretKey       string

	// Reply in group messages from manager
	onMessageReply MessageReplyCallback

	// Comment on the wall
	onWallReply MessageJsonCallback
}

type MessageReplyCallback func(from, to int)
type MessageJsonCallback func(map[string]interface{})

func (ws *WebServer) Listen(host string) {
	server := &fasthttp.Server{
		Handler:           ws.handleRequest,
		ReduceMemoryUsage: true,
	}

	log.Printf("Starting server on http://%s", host)
	err := server.ListenAndServe(host)

	if err != nil {
		log.Fatalf("error in fasthttp server: %s", err)
	}
}

func (ws *WebServer) handleRequest(ctx *fasthttp.RequestCtx) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Logger().Printf("panic when proxying the request: %s", r)
			ctx.Response.Reset()
			ctx.SetStatusCode(500)
			ctx.SetBodyString("500 Internal Server Error")
		}
	}()

	if !bytes.Equal(ctx.Method(), methodPost) {
		ctx.SetStatusCode(403)
		ctx.SetBodyString("Only POST allowed")
		return
	}

	req := &ctx.Request
	log.Println(string(req.Body()))

	var parsed map[string]interface{}
	if err := json.Unmarshal(req.Body(), &parsed); err == nil {
		if ws.secretKey != "" && ws.secretKey != parsed["secret"] {
			ctx.SetStatusCode(403)
			ctx.SetBodyString("Invalid secret")
			return
		}

		ctx.SetStatusCode(200)
		ctx.SetBodyString("ok")

		if parsed["type"] == "confirmation" {
			ctx.SetBodyString(ws.confirmationKey)
		} else if parsed["type"] == "message_reply" && ws.onMessageReply != nil {
			message := parsed["object"].(map[string]interface{})
			ws.onMessageReply(
				int(message["from_id"].(float64)),
				int(message["user_id"].(float64)),
			)
		} else if parsed["type"] == "wall_reply_new" && ws.onWallReply != nil {
			ws.onWallReply(parsed["object"].(map[string]interface{}))
		}
	}
	ctx.SetStatusCode(200)
}
