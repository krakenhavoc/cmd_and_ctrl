package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rav_b2_helpers.go — shared bodies for Ravnica: City of Guilds, batch 2
// (#5600).

// preventNextOneDamageToTheTarget is "Prevent the next 1 damage that would
// be dealt to any target this turn" as an activated ability's body
// (Benevolent Ancestor, Caregiver): preventNextDamageToTheTarget at 1.
func preventNextOneDamageToTheTarget(name string) Effect {
	return preventNextDamageToTheTarget(name, 1)
}

// preventNextDamageToTheTarget is "Prevent the next N damage that would be
// dealt to any target this turn" (Aven Redeemer's 2): the charged shield
// Kitsune Healer and Mending Hands use (PreventNextDamage, CR 615.7), on
// the one target. A target gone by resolution gets no shield (CR 608.2b).
func preventNextDamageToTheTarget(name string, n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		label := name + ": prevent the next " + strconv.Itoa(n) + " damage"
		return PreventNextDamage{Target: ts[0].ID, Amount: n, Label: label}.Apply(ctx)
	}
}
