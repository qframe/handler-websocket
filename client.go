package qhandler_websocket

import (
	"time"
	"log"

	"golang.org/x/net/websocket"
	"github.com/qnib/qframe-types"
)

// Inspired by https://github.com/golang-samples/websocket

var maxId int = 0

// Chat client.
type Client struct {
	id     int
	ws     *websocket.Conn
	plugin *Plugin
	doneCh chan bool
}

// Create new chat client.
func NewClient(ws *websocket.Conn, p *Plugin) *Client {

	if ws == nil {
		panic("ws cannot be nil")
	}

	maxId++
	doneCh := make(chan bool)

	return &Client{maxId, ws, p,doneCh}
}

func (c *Client) Conn() *websocket.Conn {
	return c.ws
}

func (c *Client) Done() {
	c.doneCh <- true
}

// Listen Write and Read request via chanel
func (c *Client) Listen() {
	c.listenWrite()
}

// Listen write request via chanel
func (c *Client) listenWrite() {
	c.plugin.Log("info", "Listening write to client")
	//bg := c.plugin.QChan.Data.Join()
	tc := c.plugin.QChan.Tick.Join()
	//lastTick := time.Now().AddDate(0,0,-1)
	lastTick := time.Now()
	for {
		select {
		// send message to the client
		case val := <-tc.Read:
			switch val.(type) {
			case qtypes.Ticker:
				tick := val.(qtypes.Ticker)
				tickDiff, _ := tick.SkipTick(lastTick)
				t := Tick{tick.Name, tickDiff.String(), tick.Duration.String()}
				log.Printf("Send via WS: %v", t)
				websocket.JSON.Send(c.ws, t)
				now := time.Now()
				lastTick = now
			}
		// receive done request
		case <-c.doneCh:
			c.doneCh <- true // for listenRead method
			return
		}
	}
}
