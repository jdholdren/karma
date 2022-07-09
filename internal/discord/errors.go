package discord

import (
	"encoding/json"
	"fmt"
	"io"
)

type errResp map[string]any

type discordErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func readErr(r io.Reader) (errResp, error) {
	var er errResp
	if err := json.NewDecoder(r).Decode(&er); err != nil {
		return errResp{}, fmt.Errorf("error decoding error: %s", err)
	}

	return er, nil
}
