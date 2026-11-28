package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wooden Sphere — Artifact {1} (Eighth Edition, #5607):
//
//	"Whenever a player casts a green spell, you may pay {1}. If you do, you
//	 gain 1 life."
//
// Treasure Nabber's watch on every player's casts and Bitter Chill's "you may
// pay" (a mana payment asked as the ability resolves; a pool or sources that
// cannot fund it degrade to a decline). The color word is read through the
// Sphere's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bd8917c-bec4-4603-bc3f-8e0c2afae56a",
		Name:         "Wooden Sphere",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB608CastASpellOfColor("G"),
				"Wooden Sphere — you may pay {1}; if you do, you gain 1 life", effects.OldB608MayPayOneToGainOne("Wooden Sphere")),
		},
	})
}
