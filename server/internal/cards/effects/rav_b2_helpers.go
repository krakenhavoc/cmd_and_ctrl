package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rav_b2_helpers.go — shared bodies for Ravnica: City of Guilds, batch 2
// (#5600).

// preventNextOneDamageToTheTarget is "Prevent the next 1 damage that would
// be dealt to any target this turn" as an activated ability's body
// (Benevolent Ancestor, Caregiver): the charged shield Kitsune Healer and
// Mending Hands use (PreventNextDamage, CR 615.7), on the one target. A
// target gone by resolution gets no shield (CR 608.2b).
func preventNextOneDamageToTheTarget(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		return PreventNextDamage{Target: ts[0].ID, Amount: 1, Label: name + ": prevent the next 1 damage"}.Apply(ctx)
	}
}
