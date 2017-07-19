package qhandler_websocket

import (
	"fmt"
	"net/http"
	"reflect"

	"golang.org/x/net/websocket"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
)


const (
	version = "0.0.1"
	pluginTyp = qtypes.HANDLER
	pluginPkg = "websocket"
)

type Plugin struct {
    qtypes.Plugin
	doneCh    chan bool
	errCh     chan error

}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	var err error
	p := Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
		doneCh: make(chan bool),
		errCh: make(chan error),
	}
	return p, err
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
	p.Log("info", "Listening server...")

	// websocket handler
	onConnected := func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				p.errCh <- err
			}
		}()
		client := NewClient(ws, p)
		client.Listen()
	}
	http.Handle("/ticker", websocket.Handler(onConnected))
	p.Log("info", "Created handler")
	for {
		select {
		case err := <-p.errCh:
			p.Log("error", fmt.Sprintf("Error: %s", err.Error()))
		case <-p.doneCh:
			return
		}
	}
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start handler %sv%s", p.Name, version))
	tick := p.CfgIntOr("ticker-msec", 2500)
	bg := p.QChan.Data.Join()
	tc := p.QChan.Tick.Join()
	p.StartTicker("websocket", tick)
	go p.Serve()
	p.Listen()
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
			p.Log("warn", fmt.Sprintf("Received Tick of type %s", reflect.TypeOf(val)))
		}
	}
}
