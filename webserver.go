package main

import (
	"encoding/json"
	"log"

	"github.com/valyala/fasthttp"
)

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
	log.Fatalln(server.ListenAndServe(host))
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

	if string(ctx.Method()) != "POST" {
		ctx.SetStatusCode(403)
		ctx.SetBodyString("Only POST allowed")
		return
	}

	req := &ctx.Request

	var parsed map[string]any
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
			message := parsed["object"].(map[string]any)
			ws.onMessageOut(
				int(message["from_id"].(float64)),
				int(message["peer_id"].(float64)),
			)
		}
	}
	ctx.SetStatusCode(200)
}
