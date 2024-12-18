package requests

type Read struct {
	Value string
}

type SetRequest struct {
	Values map[string]string `json:"values"`
}

type BroadcastRequest struct {
	Values    map[string]string `json:"values"`
	Timestamp []int             `json:"timestamp"`
}
