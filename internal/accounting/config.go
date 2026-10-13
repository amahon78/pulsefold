package accounting

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const maxBillingConfigBytes = 1 << 20

// LoadModelJSON loads a bounded, strict JSON billing model. Configuration
// errors are deliberately returned without echoing its contents because a
// command-line configuration file must not become a diagnostic leak path.
func LoadModelJSON(reader io.Reader) (Model, error) {
	if reader == nil {
		return Model{}, errors.New("billing configuration reader is nil")
	}
	payload, err := io.ReadAll(io.LimitReader(reader, maxBillingConfigBytes+1))
	if err != nil || len(payload) > maxBillingConfigBytes {
		return Model{}, errors.New("billing configuration is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var model Model
	if err := decoder.Decode(&model); err != nil {
		return Model{}, errors.New("billing configuration is invalid")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Model{}, errors.New("billing configuration is invalid")
	}
	if err := model.Validate(); err != nil {
		return Model{}, errors.New("billing configuration is invalid")
	}
	return model, nil
}
