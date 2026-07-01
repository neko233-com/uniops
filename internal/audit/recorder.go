package audit

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Recorder struct {
	sessionID uint
	commands  []RecordEntry
}

type RecordEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // input, output
	Data      string    `json:"data"`
}

func NewRecorder(sessionID uint) *Recorder {
	return &Recorder{sessionID: sessionID}
}

func (r *Recorder) RecordInput(data string) {
	r.commands = append(r.commands, RecordEntry{
		Timestamp: time.Now(),
		Type:      "input",
		Data:      data,
	})
}

func (r *Recorder) RecordOutput(data string) {
	r.commands = append(r.commands, RecordEntry{
		Timestamp: time.Now(),
		Type:      "output",
		Data:      data,
	})
}

func (r *Recorder) GetReplay() string {
	data, _ := json.Marshal(r.commands)
	return base64.StdEncoding.EncodeToString(data)
}
