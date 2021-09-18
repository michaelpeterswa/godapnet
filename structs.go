package godapnet

type Message struct {
	Text              string   `json:"text"`
	CallsignNames     []string `json:"callSignNames"`
	TransmitterGroups []string `json:"transmitterGroups"`
	Emergency         bool     `json:"emergency"`
}
