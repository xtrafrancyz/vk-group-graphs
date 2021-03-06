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
	onMessageOut MessageOutCallback
}

type MessageOutCallback func(from, to int)

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
			ctx.Logger().Printf("panic on webhook: %s", r)
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
	//log.Println(string(req.Body()))

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
		} else if parsed["type"] == "message_reply" && ws.onMessageOut != nil {
			message := parsed["object"].(map[string]interface{})
			ws.onMessageOut(
				int(message["from_id"].(float64)),
				int(message["peer_id"].(float64)),
			)
		}
	}
	ctx.SetStatusCode(200)
}
