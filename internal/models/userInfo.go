package models

import "encoding/json"

type UserInfo struct {
	Coins         int             `json:"coins"`
	InventoryJSON json.RawMessage `json:"inventory"`
	ReceivedJSON  json.RawMessage `json:"received"`
	SentJSON      json.RawMessage `json:"sent"`
}
