package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brassclaw Orcs — Creature — Orc {2}{R}, 3/2 (Fallen Empires, S82 OLD b855,
// #5637):
//
//	"This creature can't block creatures with power 2 or greater."
//
// Zurgo Bellstriker's block rule: judged against the attacker's power as blockers are
// declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8ce324f-aa58-48f9-9f56-435e49dfff0a",
		Name:         "Brassclaw Orcs",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerGE(2))),
				"it can't block creatures with power 2 or greater"),
		},
	})
}
