package helper

import (
	"encoding/base64"
	"encoding/json"

	"api-buku-kas/app/model"
)

type cursorPayload struct {
	D int64 `json:"d"`
	I int   `json:"i"`
}

func EncodeCursor(cursor model.Cursor) string {
	raw, _ := json.Marshal(cursorPayload{
		D: cursor.Date.UnixMicro(),
		I: cursor.ID,
	})

	return base64.RawURLEncoding.EncodeToString(raw)
}

func DecodeCursor(token string) (*model.Cursor, error) {
	if token == "" {
		return nil, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, BadRequest("cursor tidak valid")
	}

	var payload cursorPayload
	if err := json.Unmarshal(raw, &payload); err != nil || payload.I < 1 {
		return nil, BadRequest("cursor tidak valid")
	}

	return &model.Cursor{
		Date: timeFromMicro(payload.D),
		ID:   payload.I,
	}, nil
}