package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Madrush Cyclops — Creature — Cyclops Warrior {1}{B}{R}{G}, 3/4 (Alara
// Reborn, #5586):
//
//	"Creatures you control have haste."
//
// Fervor's layer 6 grant (CR 613.1f) to every creature you control, the
// Cyclops included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f45a2a66-04cc-4622-9c70-f2dfc1391bb1",
		Name:         "Madrush Cyclops",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "haste"),
		},
	})
}
