package qhandler_websocket

type Tick struct {
	Name 		string `json:"name"`
	Diff 		string `json:"time_diff"`
	Duration	string `json:"time_duration"`
}

