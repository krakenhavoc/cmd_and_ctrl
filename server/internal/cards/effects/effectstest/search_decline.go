package effectstest

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// search_decline_test.go — #4134, CR 701.23: a "you may search" can be
// declined outright. Declining is not "find nothing": it skips the look,
// the shuffle, EventSearchLibrary (so "whenever a player searches"
// triggers stay quiet) and an Opposition Agent's control.

// answerSearchApply answers a search prompt with {apply: v} through the
// actions layer, as `caller` for `player`.
func AnswerSearchApply(g *game.Game, c *game.PendingChoice, caller, player uuid.UUID, v bool) error {
	raw, _ := json.Marshal(map[string]any{"choice_id": c.ID.String(), "apply": v})
	return actions.Dispatch(g, actions.Action{Type: actions.TypeResolveChoice, Player: player, Caller: caller, Params: raw})
}

func LibraryIDs(p *game.Player) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(p.Library.Cards))
	for _, c := range p.Library.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func SameIDList(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
