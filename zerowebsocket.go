package zerowebsocket

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type ZeroWebSocket struct {
	wsPath    string
	eventList WebsocketEvents
}

type WebsocketEvents map[string]func(WebsocketCtx)

type WebsocketEventMessage struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type WebsocketCtx struct {
	ctx    context.Context
	svcCtx interface{}
	event  string
	conn   *websocket.Conn
	Data   interface{}
}

func New(path string) *ZeroWebSocket {
	return &ZeroWebSocket{
		wsPath:    path,
		eventList: make(WebsocketEvents),
	}
}

func (z *ZeroWebSocket) On(eventName string, handler func(ctx WebsocketCtx)) {
	z.eventList[eventName] = handler
}

func (z *ZeroWebSocket) Route(svcCtx interface{}) rest.Route {
	return rest.Route{
		Method: http.MethodGet,
		Path:   z.wsPath,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := &websocket.Upgrader{}
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logx.Error("can not upgrade websocket")
				return
			}
			defer func() {
				logx.Error("closing connection")
				c.Close()
			}()
			for {
				_, descMessage, err := c.ReadMessage()
				if err != nil {
					logx.Error(err)
					break
				}
				var websocketEventMessage WebsocketEventMessage
				if err := json.Unmarshal([]byte(string(descMessage)), &websocketEventMessage); err != nil {
					logx.Error(err)
					return
				}
				z.eventList[websocketEventMessage.Event](WebsocketCtx{
					ctx:    r.Context(),
					svcCtx: svcCtx,
					event:  websocketEventMessage.Event,
					conn:   c,
					Data:   websocketEventMessage.Data,
				})
			}
		}),
	}
}
