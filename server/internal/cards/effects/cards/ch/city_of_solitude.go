package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// City of Solitude — Enchantment {2}{G} (Visions, S82 OLD, #5631):
//
//	"Players can cast spells and activate abilities only during their own
//	 turns."
//
// Dosan the Falling Leaf's cast bar, plus an activation bar with the same
// test: the activator must be the active player. Mana abilities are not
// exempt, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "671458ac-7441-4332-9e59-a42ec8b62f31",
		Name:         "City of Solitude",
		Completeness: effects.CompletenessFull,
		CastTimings:  []game.CastTimingRule{effects.PlayersCastOnlyOnTheirOwnTurns()},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.S82OldB788PlayersActivateOnlyOnTheirOwnTurns("Players can activate abilities only during their own turns."),
		},
	})
}
