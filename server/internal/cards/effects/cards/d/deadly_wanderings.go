package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadly Wanderings — Enchantment {3}{B}{B} (Dragons of Tarkir, #5551):
//
//	"As long as you control exactly one creature, that creature gets +2/+0 and has deathtouch and lifelink."
//
// Two layered statics over the same set: the creature you control while it
// is the only one you control (a layer 7c bonus and a layer 6 grant). The
// count is read every time layers are recomputed, so a second creature
// switches both off at once (CR 611.3a).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "06d63395-6d05-44d8-bae6-0155f8026e57",
		Name:         "Deadly Wanderings",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(deadlyWanderingsOnlyCreature, 2, 0),
			effects.B16GrantKeywords(deadlyWanderingsOnlyCreature, "deathtouch", "lifelink"),
		},
	})
}

// deadlyWanderingsOnlyCreature is "that creature": the source controller's
// creature, while it is the only creature they control.
func deadlyWanderingsOnlyCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	if !target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == source.Controller && c.IsCreature() {
			n++
		}
	}
	return n == 1
}
