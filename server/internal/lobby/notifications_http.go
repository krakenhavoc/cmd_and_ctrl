package lobby

// notifications_http.go is ADR 0162 §5's GET and PUT
// /me/notifications: a signed-in person's three DM notification
// switches, over internal/notify.
//
// The feature needs the bot token (CMDCTRL_DISCORD_BOT_TOKEN), the
// public origin the link points at (CMDCTRL_PUBLIC_BASE_URL) and the
// database. Missing any of them, both routes answer 503 with
// "available": false, and the client hides the Notifications section.

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/notify"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/ratelimit"
)

// maxNotificationsRequestBytes caps a PUT body: three booleans.
const maxNotificationsRequestBytes = 1024

// notificationsResponse is the body of both routes. A 503 carries Error
// and "available": false and nothing else that means anything.
type notificationsResponse struct {
	Error string `json:"error,omitempty"`
	// Available is whether this server sends DM notifications at all.
	Available   bool `json:"available"`
	Turn        bool `json:"turn"`
	TableReady  bool `json:"table_ready"`
	TableFilled bool `json:"table_filled"`
	// DMFailed is whether Discord refused the last DM (code 50007).
	// Nothing is sent while it is true; turning a switch on clears it.
	DMFailed bool `json:"dm_failed"`
}

func notificationsResponseOf(p notify.Prefs) notificationsResponse {
	return notificationsResponse{
		Available:   true,
		Turn:        p.Turn,
		TableReady:  p.TableReady,
		TableFilled: p.TableFilled,
		DMFailed:    !p.DMFailedAt.IsZero(),
	}
}

// putNotificationsRequest is the body of PUT /me/notifications. A
// switch left out keeps its value; at least one must be named.
type putNotificationsRequest struct {
	Turn        *bool `json:"turn"`
	TableReady  *bool `json:"table_ready"`
	TableFilled *bool `json:"table_filled"`
}

// notificationsAvailable is ADR 0162 §4's "no bot token: the feature
// is off": the token, an origin for the link, and a store to keep the
// switches in.
func (c Config) notificationsAvailable() bool {
	return c.DiscordBot.Enabled() && c.InviteBaseURL != "" && c.Notifications != nil
}

// writeNotificationsOff is the 503 both routes answer while the
// feature is off.
func writeNotificationsOff(w http.ResponseWriter) error {
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusServiceUnavailable, notificationsResponse{
		Error:     "Discord notifications are not available on this server",
		Available: false,
	})
}

// myNotifications is GET /me/notifications. Same caller rule as the
// rest of /me/*: a signed-in person, so a guest and the admin token get
// 403 (never 401, #1154).
func myNotifications(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	if !c.notificationsAvailable() {
		return writeNotificationsOff(w)
	}
	prefs, err := c.Notifications.Get(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("reading a user's notification switches failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your notification settings; try again")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, notificationsResponseOf(prefs))
}

// putMyNotifications is PUT /me/notifications. Per-person bucket: 1 a
// second, a burst of 5, spent only by a write that passed its checks.
func putMyNotifications(limit *ratelimit.Limiter) lobbyHandler {
	return func(c Config, w http.ResponseWriter, r *http.Request) error {
		p, err := signedInUser(r)
		if err != nil {
			return err
		}
		if !c.notificationsAvailable() {
			return writeNotificationsOff(w)
		}
		var req putNotificationsRequest
		r.Body = http.MaxBytesReader(w, r.Body, maxNotificationsRequestBytes)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			return httpError(http.StatusBadRequest, "invalid body: "+err.Error())
		}
		u := notify.Update{Turn: req.Turn, TableReady: req.TableReady, TableFilled: req.TableFilled}
		if u.Empty() {
			return httpError(http.StatusBadRequest, "name at least one of turn, table_ready, table_filled")
		}
		if !limit.Allow("user:" + p.UserID.String()) {
			w.Header().Set("Retry-After", "1")
			return httpError(http.StatusTooManyRequests, "too many changes; try again in a second")
		}
		saved, err := c.Notifications.Put(r.Context(), p.UserID, u)
		switch {
		case err == nil:
			w.Header().Set("Cache-Control", "no-store")
			return writeJSON(w, http.StatusOK, notificationsResponseOf(saved))
		case errors.Is(err, notify.ErrNoStore):
			return writeNotificationsOff(w)
		default:
			c.logger().Error("saving a user's notification switches failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not save your notification settings; try again")
		}
	}
}
