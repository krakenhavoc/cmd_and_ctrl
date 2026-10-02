package lobby

// settings_http.go is ADR 0110 section 4's account settings: GET and
// PUT /me/settings (Delivery PR 5), over internal/usersettings.
//
// The body is the synced subset of the client's Settings: the
// per-person fields (SYNCED_FIELDS in client/src/lib/settings.ts). The
// per-device ones never leave the browser. The server checks only the
// body's shape; the client's migrate stays the one schema validator.

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/usersettings"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// maxSettingsRequestBytes caps the whole PUT request: the settings
// body's own cap plus room for the envelope. Anything past it is cut
// off at the reader, before it is decoded.
const maxSettingsRequestBytes = usersettings.MaxBodyBytes + 1024

// settingsResponse is the body of GET /me/settings, of a successful
// PUT, and of the 412 and 409 refusals (which add Error). A user with
// no copy yet gets {"revision": 0} and nothing else.
type settingsResponse struct {
	Error     string          `json:"error,omitempty"`
	Version   int             `json:"version,omitempty"`
	Revision  int64           `json:"revision"`
	Settings  json.RawMessage `json:"settings,omitempty"`
	UpdatedAt int64           `json:"updated_at,omitempty"`
}

func settingsResponseOf(s usersettings.Settings) settingsResponse {
	out := settingsResponse{Version: s.Version, Revision: s.Revision, Settings: s.Body}
	if !s.UpdatedAt.IsZero() {
		out.UpdatedAt = s.UpdatedAt.UnixMilli()
	}
	return out
}

// putSettingsRequest is the body of PUT /me/settings.
type putSettingsRequest struct {
	// Version is the client's SETTINGS_VERSION.
	Version int `json:"version"`
	// Settings is the synced subset, a JSON object.
	Settings json.RawMessage `json:"settings"`
}

func (c Config) userSettings() usersettings.Store {
	if c.UserSettings == nil {
		return usersettings.NoStore{}
	}
	return c.UserSettings
}

// mySettings is GET /me/settings: the caller's account copy. Same
// caller rule as the rest of /me/*: a signed-in person, so a guest,
// the admin token and a deployment with no database get 403 (never
// 401, #1154), which the client reads as "browser-only".
func mySettings(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	cur, err := c.userSettings().Get(r.Context(), p.UserID)
	if errors.Is(err, usersettings.ErrNotFound) {
		return writeSettings(w, http.StatusOK, settingsResponse{Revision: 0})
	}
	if err != nil {
		c.logger().Error("reading a user's settings failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your settings; try again")
	}
	return writeSettings(w, http.StatusOK, settingsResponseOf(cur))
}

// putMySettings is PUT /me/settings, with If-Match: <revision>.
//
//   - 428 with no If-Match, 400 for one that is not a revision.
//   - 400 for a malformed body, 413 for settings over 32 KiB.
//   - 429 past the per-person bucket (1 a second, a burst of 5). Only a
//     write that would reach the database spends a token.
//   - 412 with the current copy when the revision has moved, so two
//     tabs or devices never silently overwrite each other.
//   - 409 with the current copy when the client's version is older
//     than the stored one: a stale tab must not stamp an older schema
//     over a newer client's copy.
func putMySettings(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		ifMatch, err := parseIfMatch(r.Header.Get("If-Match"))
		if err != nil {
			return err
		}
		req, err := decodeSettingsRequest(w, r)
		if err != nil {
			return err
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "1")
			return httpError(http.StatusTooManyRequests, "too many settings writes; try again in a second")
		}
		saved, err := c.userSettings().Put(r.Context(), p.UserID, req.Version, req.Settings, ifMatch)
		switch {
		case err == nil:
			return writeSettings(w, http.StatusOK, settingsResponseOf(saved))
		case errors.Is(err, usersettings.ErrRevisionConflict):
			out := settingsResponseOf(saved)
			out.Error = "these settings were changed somewhere else"
			return writeSettings(w, http.StatusPreconditionFailed, out)
		case errors.Is(err, usersettings.ErrVersionTooOld):
			out := settingsResponseOf(saved)
			out.Error = "a newer version of the site saved these settings; reload"
			return writeSettings(w, http.StatusConflict, out)
		case errors.Is(err, usersettings.ErrInvalid):
			return httpError(http.StatusBadRequest, validationMessage(err))
		case errors.Is(err, usersettings.ErrNoStore):
			return httpError(http.StatusServiceUnavailable, "settings cannot be saved on this server")
		default:
			c.logger().Error("saving a user's settings failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not save your settings; try again")
		}
	}
}

// parseIfMatch reads the revision a PUT is conditional on. A bare
// integer and a quoted entity tag ("3", W/"3") are both accepted.
func parseIfMatch(v string) (int64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, httpError(http.StatusPreconditionRequired, "If-Match is required: send the revision you last read (0 for a first copy)")
	}
	v = strings.TrimPrefix(v, "W/")
	v = strings.Trim(v, `"`)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return 0, httpError(http.StatusBadRequest, "If-Match must be a settings revision")
	}
	return n, nil
}

// decodeSettingsRequest decodes and checks a PUT body before any
// rate-limit token or database work is spent on it.
func decodeSettingsRequest(w http.ResponseWriter, r *http.Request) (putSettingsRequest, error) {
	var req putSettingsRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxSettingsRequestBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return req, httpError(http.StatusRequestEntityTooLarge, "settings are larger than 32 KiB")
		}
		return req, httpError(http.StatusBadRequest, "invalid body: "+err.Error())
	}
	if len(req.Settings) == 0 {
		return req, httpError(http.StatusBadRequest, "settings is required")
	}
	// Stored compact, so the cap measures the settings and not the
	// sender's whitespace.
	var compact bytes.Buffer
	if err := json.Compact(&compact, req.Settings); err != nil {
		return req, httpError(http.StatusBadRequest, "settings must be a JSON object")
	}
	req.Settings = compact.Bytes()
	if len(req.Settings) > usersettings.MaxBodyBytes {
		return req, httpError(http.StatusRequestEntityTooLarge, "settings are larger than 32 KiB")
	}
	if err := usersettings.Validate(req.Version, req.Settings); err != nil {
		return req, httpError(http.StatusBadRequest, validationMessage(err))
	}
	return req, nil
}

// validationMessage is the player-facing half of a usersettings
// refusal: Validate joins ErrInvalid with the reason, one per line.
func validationMessage(err error) string {
	msg := err.Error()
	if i := strings.LastIndex(msg, "\n"); i >= 0 {
		msg = msg[i+1:]
	}
	return msg
}

// writeSettings is writeJSON for a person's own settings: never cached
// by the browser or anything between, since the copy changes under
// the same URL.
func writeSettings(w http.ResponseWriter, status int, body settingsResponse) error {
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, status, body)
}
