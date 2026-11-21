package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravitational Shift — Enchantment {3}{U}{U} (Rise of the Eldrazi, #5581):
//
//	"Creatures with flying get +2/+0.
//	 Creatures without flying get -2/-0."
//
// Crosswinds' layer 7c reading of flying, twice: every creature on the
// battlefield gets one of the two, read after layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0411aead-9f77-435d-a343-635f23221cf7",
		Name:         "Gravitational Shift",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Labelled(effects.B16Anthem(effects.CrosswindsCreaturesWithFlying, 2, 0), "Creatures with flying get +2/+0."),
			effects.Labelled(effects.B16Anthem(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature() && !game.HasKeyword(target, "flying")
			}, -2, 0), "Creatures without flying get -2/-0."),
		},
	})
}
