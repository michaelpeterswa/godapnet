package godapnet

type Sender struct {
	Callsign string
	Username string
	Password string
}

type Message struct {
	Text                  string   `json:"text"`
	CallsignNames         []string `json:"callSignNames"`
	TransmitterGroupNames []string `json:"transmitterGroupNames"`
	Emergency             bool     `json:"emergency"`
}
