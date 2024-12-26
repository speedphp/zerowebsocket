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
	Conn *websocket.Conn

	ZeroWebSocket struct {
		wsPath    string
		eventList WebsocketEvents
	}

	Event struct {
		Event   string
		Handler EventHandler
	}

	WebsocketEvents map[string]EventHandler

	ConnectedHandler func(WebsocketCtx) interface{}

	OriginHandler func(*http.Request) bool

	CloseHandler func(WebsocketCtx) error

	ErrorHandler func(WebsocketCtx, error)

	EventHandler func(WebsocketCtx)

	WebsocketEventMessage struct {
		Event string      `json:"event"`
		Data  interface{} `json:"data"`
	}

	WebsocketCtx struct {
		Ctx           context.Context
		SvcCtx        interface{}
		Event         string
		Conn          *websocket.Conn
		Data          interface{}
		Req           *http.Request
		ConnectedInfo interface{}
	}

	RouteOptions struct {
		SvcCtx           interface{}
		OriginHandler    OriginHandler
		CloseHandler     CloseHandler
		ErrorHandler     ErrorHandler
		ConnectedHandler ConnectedHandler
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

func (z *ZeroWebSocket) Route(opts *RouteOptions) rest.Route {
	if opts.OriginHandler == nil {
		opts.OriginHandler = func(r *http.Request) bool {
			return true
		}
	}
	if opts.ErrorHandler == nil {
		opts.ErrorHandler = func(wc WebsocketCtx, err error) {
			logx.Error(err)
		}
	}
	return rest.Route{
		Method: http.MethodGet,
		Path:   z.wsPath,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			upgrader := &websocket.Upgrader{
				CheckOrigin: opts.OriginHandler,
			}
			c, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logx.Error("can not upgrade websocket")
				return
			}
			var ConnectedInfo interface{}
			if opts.ConnectedHandler != nil {
				ConnectedInfo = opts.ConnectedHandler(WebsocketCtx{
					Ctx:           r.Context(),
					SvcCtx:        opts.SvcCtx,
					Event:         "",
					Conn:          c,
					Data:          nil,
					Req:           r,
					ConnectedInfo: nil,
				})
			}
			defaultCtx := WebsocketCtx{
				Ctx:           r.Context(),
				SvcCtx:        opts.SvcCtx,
				Event:         "",
				Conn:          c,
				Data:          nil,
				Req:           r,
				ConnectedInfo: ConnectedInfo,
			}
			defer func() {
				if opts.CloseHandler != nil {
					opts.CloseHandler(defaultCtx)
				}
				c.Close()
			}()
			for {
				_, descMessage, err := c.ReadMessage()
				if err != nil {
					opts.ErrorHandler(defaultCtx, err)
					break
				}
				var websocketEventMessage WebsocketEventMessage
				if err := json.Unmarshal([]byte(string(descMessage)), &websocketEventMessage); err != nil {
					opts.ErrorHandler(defaultCtx, err)
					return
				}
				z.eventList[websocketEventMessage.Event](WebsocketCtx{
					Ctx:           r.Context(),
					SvcCtx:        opts.SvcCtx,
					Event:         websocketEventMessage.Event,
					Conn:          c,
					Data:          websocketEventMessage.Data,
					Req:           r,
					ConnectedInfo: ConnectedInfo,
				})
			}
		}),
	}
}
