package zerowebsocket

import (
	"encoding/json"
	"net/http"

	"context"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

type (
	ZeroWebSocket struct {
		wsPath    string
		eventList WebsocketEvents
	}

	Event struct {
		Event   string
		Handler EventHandler
	}

	WebsocketEvents map[string]EventHandler

	OriginHandler func(*http.Request) bool

	EventHandler func(WebsocketCtx)

	WebsocketEventMessage struct {
		Event string      `json:"event"`
		Data  interface{} `json:"data"`
	}

	WebsocketCtx struct {
		Ctx    context.Context
		SvcCtx interface{}
		Event  string
		Conn   *websocket.Conn
		Data   interface{}
	}
)

const (
	// Reproduced from the gorilla/websocket package
	// so you don't have to import it again
	TextMessage   = 1
	BinaryMessage = 2
	CloseMessage  = 8
	PingMessage   = 9
	PongMessage   = 10
)

func New(path string) *ZeroWebSocket {
	return &ZeroWebSocket{
		wsPath:    path,
		eventList: make(WebsocketEvents),
	}
}

func (z *ZeroWebSocket) On(eventName string, handler func(ctx WebsocketCtx)) {
	z.eventList[eventName] = handler
}

func OnEvents(z *ZeroWebSocket, events ...Event) {
	for _, e := range events {
		z.eventList[e.Event] = e.Handler
	}
}

func (z *ZeroWebSocket) RouteWithOrigin(svcCtx interface{}, originHandler OriginHandler) rest.Route {
	if originHandler == nil {
		originHandler = func(r *http.Request) bool {
			return true
		}
	}
	return rest.Route{
		Method: http.MethodGet,
		Path:   z.wsPath,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := &websocket.Upgrader{
				CheckOrigin: originHandler,
			}
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
					Ctx:    r.Context(),
					SvcCtx: svcCtx,
					Event:  websocketEventMessage.Event,
					Conn:   c,
					Data:   websocketEventMessage.Data,
				})
			}
		}),
	}
}

func (z *ZeroWebSocket) Route(svcCtx interface{}) rest.Route {
	return z.RouteWithOrigin(svcCtx, nil)
}
