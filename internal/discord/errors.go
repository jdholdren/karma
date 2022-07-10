package discord

import (
	"encoding/json"
	"fmt"
	"io"
)

type errResp map[string]any

func (er errResp) Error() string {
	return fmt.Sprintf("%#v", er)
}

func readErr(r io.Reader) (errResp, error) {
	var er errResp
	if err := json.NewDecoder(r).Decode(&er); err != nil {
		return errResp{}, fmt.Errorf("error decoding error: %s", err)
	}

	return er, nil
}
