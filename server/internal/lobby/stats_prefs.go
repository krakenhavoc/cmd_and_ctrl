package lobby

// stats_prefs.go is ADR 0217 §2 and owner answer 1 (slice 3 of #4880):
// the advanced-stats setting.
//
//   - GET and PUT /me/stats/prefs: a signed-in person's own switch.
//     A guest and the admin token are a 403 (signedInUser): guests get
//     no stats.
//   - PUT /admin/users/{id}/advanced-stats: an admin switches it for a
//     person. The person can still switch it off.
//   - GET /admin/stats/advanced: the server's default and every account
//     whose switch is set, with who set it. Every other account follows
//     the default.
//   - PUT /admin/stats/advanced-default: an admin stores the default
//     for everyone who never touched their switch. No stored default is
//     gamestats.AdvancedDefault (off).
//
// The admin routes ride requireAdmin, whose "admin action" line records
// the route and who called it; each write adds one more "admin action"
// line naming what it set. None of them reads anyone's stats.

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
)

// errNoStatsPrefs is the writes' answer on a server with no database.
var errNoStatsPrefs = httpError(http.StatusServiceUnavailable, "the stats setting needs the user database, and this server has none")

// Choice names a person's switch: "on" or "off" when they or an admin
// set it, "default" when it follows the server's default.
const (
	statsChoiceOn      = "on"
	statsChoiceOff     = "off"
	statsChoiceDefault = "default"
)

func statsChoice(p gamestats.Prefs) string {
	switch {
	case p.Advanced == nil:
		return statsChoiceDefault
	case *p.Advanced:
		return statsChoiceOn
	default:
		return statsChoiceOff
	}
}

func millisOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

// statsPrefsResponse is GET and PUT /me/stats/prefs' body.
type statsPrefsResponse struct {
	// Available is false on a server with no database: nothing can be
	// switched, and advanced is the code default.
	Available bool `json:"available"`
	// Advanced is the answer: whether advanced stats are on for the
	// caller now.
	Advanced bool `json:"advanced"`
	// Choice is "on", "off" or "default".
	Choice string `json:"choice"`
	// SetBy is "self" or "admin" when Choice is not "default".
	SetBy string `json:"set_by,omitempty"`
	// Default is the server's default, which "default" follows.
	Default bool `json:"default"`
	// ShareLinkAt is when the caller's share link was made (ADR 0217
	// §6), Unix ms, absent while they have none. The link itself is
	// shown once, when it is made: only its hash is kept.
	ShareLinkAt int64 `json:"share_link_at,omitempty"`
	// SharePods is whether the caller shares their stats with the
	// members of every pod they are in (ADR 0220 §7): one switch for all
	// pods, off until they turn it on.
	SharePods bool `json:"share_pods"`
}

// advancedStatsFor is userID's setting and the server's default. A
// server with no store has neither: the person follows
// gamestats.AdvancedDefault.
func (c Config) advancedStatsFor(ctx context.Context, userID uuid.UUID) (gamestats.Prefs, gamestats.DefaultSetting, error) {
	if c.StatsPrefs == nil {
		return gamestats.Prefs{}, gamestats.DefaultSetting{Advanced: gamestats.AdvancedDefault}, nil
	}
	def, err := c.StatsPrefs.Default(ctx)
	if err != nil {
		return gamestats.Prefs{}, gamestats.DefaultSetting{}, err
	}
	p, err := c.StatsPrefs.Prefs(ctx, userID.String())
	if err != nil {
		return gamestats.Prefs{}, gamestats.DefaultSetting{}, err
	}
	return p, def, nil
}

func (c Config) statsPrefsResponseOf(p gamestats.Prefs, def gamestats.DefaultSetting) statsPrefsResponse {
	return statsPrefsResponse{
		Available:   c.StatsPrefs != nil,
		Advanced:    p.Effective(def.Advanced),
		Choice:      statsChoice(p),
		SetBy:       string(p.SetBy),
		Default:     def.Advanced,
		ShareLinkAt: millisOrZero(p.ShareLinkAt),
		SharePods:   p.SharePods,
	}
}

// myStatsPrefs is GET /me/stats/prefs.
func myStatsPrefs(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	prefs, def, err := c.advancedStatsFor(r.Context(), p.UserID)
	if err != nil {
		c.logger().Error("reading a user's stats setting failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your stats setting; try again")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, c.statsPrefsResponseOf(prefs, def))
}

// putStatsPrefsRequest is PUT /me/stats/prefs' body: either switch or
// both. A pointer, so a missing field is left alone, and a body with
// neither is a 400, never a silent "off".
type putStatsPrefsRequest struct {
	Advanced  *bool `json:"advanced"`
	SharePods *bool `json:"share_pods"`
}

// putMyStatsPrefs is PUT /me/stats/prefs: the person switches advanced
// stats on or off for themselves, which also overrides an admin's
// switch and the default from now on.
func putMyStatsPrefs(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	var body putStatsPrefsRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Advanced == nil && body.SharePods == nil {
		return httpError(http.StatusBadRequest, `body must be {"advanced": true|false}, {"share_pods": true|false}, or both`)
	}
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	var saved gamestats.Prefs
	save := func(set func() (gamestats.Prefs, error)) error {
		got, err := set()
		if errors.Is(err, gamestats.ErrNoUser) {
			return httpError(http.StatusNotFound, "user not found")
		}
		if err != nil {
			c.logger().Error("saving a user's stats setting failed", "err", err)
			return httpError(http.StatusInternalServerError, "could not save your stats setting; try again")
		}
		saved = got
		return nil
	}
	id := p.UserID.String()
	if body.Advanced != nil {
		if err := save(func() (gamestats.Prefs, error) {
			return c.StatsPrefs.SetAdvanced(r.Context(), id, *body.Advanced, gamestats.SetBySelf, c.now())
		}); err != nil {
			return err
		}
	}
	if body.SharePods != nil {
		if err := save(func() (gamestats.Prefs, error) {
			return c.StatsPrefs.SetSharePods(r.Context(), id, *body.SharePods, c.now())
		}); err != nil {
			return err
		}
	}
	def, err := c.StatsPrefs.Default(r.Context())
	if err != nil {
		c.logger().Error("reading the advanced stats default failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your stats setting; try again")
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, c.statsPrefsResponseOf(saved, def))
}

// adminSwitchRequest is the admin writes' body, as PUT /me/admin-mode's.
type adminSwitchRequest struct {
	Enabled *bool `json:"on"`
}

func decodeAdminSwitch(w http.ResponseWriter, r *http.Request) (bool, error) {
	var body adminSwitchRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return false, err
	}
	if body.Enabled == nil {
		return false, httpError(http.StatusBadRequest, `body must be {"on": true} or {"on": false}`)
	}
	return *body.Enabled, nil
}

// adminAccountStats is one account's line in the admin listing.
type adminAccountStats struct {
	UserID string `json:"user_id"`
	// Choice is "on", "off" or "default".
	Choice string `json:"choice"`
	// SetBy is "self" or "admin"; empty while Choice is "default".
	SetBy string `json:"set_by,omitempty"`
	// Advanced is the answer for this person now.
	Advanced  bool  `json:"advanced"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

func adminAccountStatsOf(userID string, p gamestats.Prefs, def bool) adminAccountStats {
	return adminAccountStats{
		UserID:    userID,
		Choice:    statsChoice(p),
		SetBy:     string(p.SetBy),
		Advanced:  p.Effective(def),
		UpdatedAt: millisOrZero(p.UpdatedAt),
	}
}

// adminDefaultResponse is the server's advanced-stats default.
type adminDefaultResponse struct {
	Advanced bool `json:"on"`
	// Stored is false while no admin has set one, and on is then the
	// code's fallback, gamestats.AdvancedDefault.
	Stored    bool  `json:"stored"`
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

func adminDefaultOf(d gamestats.DefaultSetting) adminDefaultResponse {
	return adminDefaultResponse{Advanced: d.Advanced, Stored: d.Stored, UpdatedAt: millisOrZero(d.UpdatedAt)}
}

// adminAdvancedStatsResponse is GET /admin/stats/advanced's body.
type adminAdvancedStatsResponse struct {
	Default adminDefaultResponse `json:"default"`
	// Accounts is every account whose switch is set, newest change
	// first. An account not listed follows the default.
	Accounts  []adminAccountStats `json:"accounts"`
	Truncated bool                `json:"truncated,omitempty"`
}

// adminAdvancedStats is GET /admin/stats/advanced.
func adminAdvancedStats(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	def, err := c.StatsPrefs.Default(r.Context())
	if err != nil {
		return adminViewFailed(c, "advanced stats", err)
	}
	rows, err := c.StatsPrefs.Accounts(r.Context(), gamestats.MaxPrefsAccounts+1)
	if err != nil {
		return adminViewFailed(c, "advanced stats", err)
	}
	out := adminAdvancedStatsResponse{Default: adminDefaultOf(def), Accounts: []adminAccountStats{}}
	if len(rows) > gamestats.MaxPrefsAccounts {
		rows, out.Truncated = rows[:gamestats.MaxPrefsAccounts], true
	}
	for _, a := range rows {
		out.Accounts = append(out.Accounts, adminAccountStatsOf(a.UserID, a.Prefs, def.Advanced))
	}
	w.Header().Set("Cache-Control", "no-store")
	return writeJSON(w, http.StatusOK, out)
}

// adminSetAdvancedStats is PUT /admin/users/{id}/advanced-stats: an
// admin switches advanced stats on or off for one person, recorded as
// set by an admin. It reads no stats.
func adminSetAdvancedStats(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil || id == uuid.Nil {
		return httpError(http.StatusBadRequest, "invalid user id")
	}
	on, err := decodeAdminSwitch(w, r)
	if err != nil {
		return err
	}
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	saved, err := c.StatsPrefs.SetAdvanced(r.Context(), id.String(), on, gamestats.SetByAdmin, c.now())
	if errors.Is(err, gamestats.ErrNoUser) {
		return httpError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		c.logger().Error("an admin's advanced stats switch failed", "user_id", id.String(), "err", err)
		return httpError(http.StatusInternalServerError, "could not save the switch; try again")
	}
	if p, ok := auth.PrincipalFromContext(r.Context()); ok {
		logAdminAction(c.Log, "advanced stats set", p, "user_id", id.String(), "on", on)
	}
	def, err := c.StatsPrefs.Default(r.Context())
	if err != nil {
		def = gamestats.DefaultSetting{Advanced: gamestats.AdvancedDefault}
	}
	return writeJSON(w, http.StatusOK, adminAccountStatsOf(id.String(), saved, def.Advanced))
}

// adminSetAdvancedDefault is PUT /admin/stats/advanced-default: the
// default for everyone whose switch is unset. Anyone who switched it
// themselves, on or off, keeps their answer.
func adminSetAdvancedDefault(c Config, w http.ResponseWriter, r *http.Request) error {
	on, err := decodeAdminSwitch(w, r)
	if err != nil {
		return err
	}
	if c.StatsPrefs == nil {
		return errNoStatsPrefs
	}
	saved, err := c.StatsPrefs.SetDefault(r.Context(), on, c.now())
	if err != nil {
		c.logger().Error("saving the advanced stats default failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not save the default; try again")
	}
	if p, ok := auth.PrincipalFromContext(r.Context()); ok {
		logAdminAction(c.Log, "advanced stats default set", p, "on", on)
	}
	return writeJSON(w, http.StatusOK, adminDefaultOf(saved))
}
