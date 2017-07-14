package qhandler_websocket

type Message struct {
	Author string `json:"author"`
	Data   map[string]interface{} `json:"data"`
}

func (self *Message) String() string {
	return self.Author + " says something"
}
