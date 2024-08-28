# zerowebsocket

简化go-zero项目的websocket服务

## 特性

- 采用[gorilla/websocket](https://github.com/gorilla/websocket)的WebSocket实现。
- 增加事件机制，WebSocket事件可以设置对应的处理Handler。
- 简单易用，给go-zero项目增加一行路由即可。

## 使用方法

```
// 1. 引入 gorilla/websocket 和 zerowebsocket
import (
  "github.com/gorilla/websocket"
  "github.com/speedphp/zerowebsocket"
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	// 2. 设置WebSocket路径
	ws := zerowebsocket.New("/ws")

	// 3. 增加WebSocket事件和对应的handler
	ws.On("test", func(ctx zerowebsocket.WebsocketCtx) {
		word := ctx.Data.(string)                                           // 接收到数据
		ctx.Conn.WriteMessage(websocket.TextMessage, []byte("hello " + word)) // 发送数据
	})

	// 4. 设置路由
	server.AddRoute(ws.Route(ctx))

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
```

## 测试

使用[websocat](https://github.com/vi/websocat)在命令行测试WebSocket联通情况：

```bash
websocat ws://localhost:8080/ws
{"event": "test", "data": "ok"}
hello ok
```

## 感谢

- [go-zero](https://github.com/zeromicro/go-zero) 是一个集成了各种工程实践的 web 和 rpc 框架。通过弹性设计保障了大并发服务端的稳定性，经受了充分的实战检验。
- [gorilla/websocket](https://github.com/gorilla/websocket) 常用的Go语言WebSocket库。
