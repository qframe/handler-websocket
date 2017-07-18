package qhandler_websocket

import (
	"io"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"golang.org/x/net/websocket"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
)

const (
	version = "0.0.1"
	pluginTyp = qtypes.HANDLER
	pluginPkg = "influxdb"
)

type Plugin struct {
    qtypes.Plugin
	ws *websocket.Conn
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	var err error
	p := Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}
	return p, err
}

func echoHandler(ws *websocket.Conn) {
	fmt.Println("huhu")
	io.Copy(ws, ws)
}

// Connect creates a connection to InfluxDB
func (p *Plugin) Serve() {
	host := p.CfgStringOr("bind-host", "")
	port := p.CfgStringOr("bind-port", "1234")
	addr := fmt.Sprintf("%s:%s", host, port)
	http.Handle("/", http.FileServer(assetFS()))
	p.Log("info", fmt.Sprintf("Start webserver on '%s'", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		p.Log("error", fmt.Sprintf("Error serving on '%s': %v", addr, err))
	}
}

func (p *Plugin) Listen() {
	server := NewServer("/entry")
	server.Listen()
}

func (p *Plugin) SendToWS() {
	tc := p.QChan.Tick.Join()
	// Initialise lastTick with time of a year ago
	lastTick := time.Now().AddDate(0,0,-1)
	for {
		select {
		case val := <-tc.Read:
			switch val.(type) {
			case qtypes.Ticker:
				tick := val.(qtypes.Ticker)
				tickDiff, _ := tick.SkipTick(lastTick)
				msg := fmt.Sprintf("tick '%s' | Last tick %s ago (< %s)", tick.Name, tickDiff.String(), tick.Duration.String())
				p.Log("info", "Send via WS: "+msg)
				websocket.Message.Send(p.ws, msg)
				now := time.Now()
				lastTick = now
			}
		}
	}
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start handler %sv%s", p.Name, version))
	tick := p.CfgIntOr("ticker-msec", 1000)
	bg := p.QChan.Data.Join()
	tc := p.QChan.Tick.Join()
	p.StartTicker("websocket", tick)
	go p.Serve()
	go p.Listen()
	// Initialise lastTick with time of a year ago
	//lastTick := time.Now().AddDate(0,0,-1)
	for {
		select {
		case val := <-bg.Read:
			switch val.(type) {
			case qtypes.Metric:
				m := val.(qtypes.Metric)
				if p.StopProcessingMetric(m, false) {
					continue
				}
			}
		case val := <-tc.Read:
			switch val.(type) {
			case qtypes.Ticker:
				continue
				/*
				tick := val.(qtypes.Ticker)
				tickDiff, skipTick := tick.SkipTick(lastTick)
				if ! skipTick {
					msg := fmt.Sprintf("tick '%s' | Last tick %s ago (< %s)", tick.Name, tickDiff.String(), tick.Duration.String())
					p.Log("info", "Send via WS: "+msg)
					err := websocket.Message.Send(p.ws, msg)
					if err != nil {
						p.Log("error", err.Error())
					}
				}
				now := time.Now()
				lastTick = now
				*/
			default:
				p.Log("warn", fmt.Sprintf("Received Tick of type %s", reflect.TypeOf(val)))
			}
		}
	}
}
