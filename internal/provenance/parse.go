package provenance

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

const maxManifestBytes = 1 << 20

// ParseManifestJSON decodes a bounded manifest representation for later
// consumers. It accepts no unknown fields, avoids echoing input and performs
// semantic validation before returning a manifest.
func ParseManifestJSON(reader io.Reader) (Manifest, error) {
	if reader == nil {
		return Manifest{}, errors.New("manifest reader is nil")
	}
	payload, err := io.ReadAll(io.LimitReader(reader, maxManifestBytes+1))
	if err != nil || len(payload) > maxManifestBytes {
		return Manifest{}, errors.New("manifest is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, errors.New("manifest is invalid")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Manifest{}, errors.New("manifest is invalid")
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, errors.New("manifest is invalid")
	}
	return manifest, nil
}
