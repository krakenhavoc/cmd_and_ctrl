package lobby

// botmodel.go is #2980's operator surface: GET and PUT /admin/bot/model,
// the readout of the hosted-model spend guard (aiseat/budget) and its
// kill switch. Admin-only like every /admin route (requireAdmin, ADR
// 0110 §3); the shared token and an allowlisted person in admin mode
// both pass, a seated player does not.
//
// The same budget feeds the player-facing half: botOptions copies
// Status().Note onto GET /bot/options as model_note, and the model
// tiers' own reasons come from the tier factory (tiers.ModelGate).

import (
	"net/http"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/budget"
)

// BotModelBudget is what the admin routes and the bot picker need of
// the spend guard. *budget.Budget satisfies it.
type BotModelBudget interface {
	Status() budget.Status
	SetEnabled(enabled bool) error
}

// botModelStatus answers GET /admin/bot/model. 503 when this server
// has no guard to report on: no model transport is configured (no
// hosted key and no local endpoint), so there is nothing to cap.
func botModelStatus(c Config, w http.ResponseWriter, _ *http.Request) error {
	if c.BotBudget == nil {
		return httpError(http.StatusServiceUnavailable, "no bot model budget on this server")
	}
	return writeJSON(w, http.StatusOK, c.BotBudget.Status())
}

// botModelSwitchRequest is PUT /admin/bot/model's body. Enabled is a
// pointer so a missing field is a 400, never a silent "off" (or "on").
type botModelSwitchRequest struct {
	Enabled *bool `json:"enabled"`
}

// botModelSwitch answers PUT /admin/bot/model: the kill switch.
// {"enabled": false} turns every model call off, for every table, at
// once and without a redeploy; seats already playing fall back to the
// heuristic on their next decision. {"enabled": true} turns them back
// on (the daily and monthly caps still apply). The setting is
// persisted. Replies with the new status.
func botModelSwitch(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.BotBudget == nil {
		return httpError(http.StatusServiceUnavailable, "no bot model budget on this server")
	}
	var body botModelSwitchRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Enabled == nil {
		return httpError(http.StatusBadRequest, "enabled is required")
	}
	if err := c.BotBudget.SetEnabled(*body.Enabled); err != nil {
		return httpError(http.StatusInternalServerError, "could not save the switch: "+err.Error())
	}
	return writeJSON(w, http.StatusOK, c.BotBudget.Status())
}
