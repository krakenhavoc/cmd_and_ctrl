package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b232_helpers.go — the cast-time reading Draconic Roar and Orator of
// Ojutai share (S82 older sets, batch 232).

// ControlsADragonAsCast is the CastTimeValues reader of "if you ... controlled
// a Dragon as you cast this spell" (Draconic Roar, Orator of Ojutai): 1 when
// the caster controls a Dragon creature as the spell is announced, read
// through the layers so a changeling counts (CR 601.2i), else 0.
func ControlsADragonAsCast(g *game.Game, item *game.StackItem) int {
	dragon := NewContext(g, item).TextSubtype("Dragon")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.HasSubtype(dragon) {
			return 1
		}
	}
	return 0
}
