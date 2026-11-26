package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scourge of Geier Reach — Creature — Elemental {3}{R}{R}, 3/3 (Innistrad,
// S82 older sets batch 449, #5572):
//
//	"This creature gets +1/+1 for each creature your opponents control."
//
// A layer-7c modification counted live, the mirror of Shanna's count of your
// own creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e02f371-5dce-415a-8563-1f93c35401d5",
		Name:         "Scourge of Geier Reach",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB449OpponentCreaturesPlusPT("This creature gets +1/+1 for each creature your opponents control."),
		},
	})
}
