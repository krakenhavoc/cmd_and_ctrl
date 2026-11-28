package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desolation — Enchantment {1}{B}{B} (Visions, S82 OLD batch 786, #5631):
//
//	"At the beginning of each end step, each player who tapped a land for
//	 mana this turn sacrifices a land of their choice. This enchantment deals
//	 2 damage to each player who sacrificed a Plains this way."
//
// Any player's end step triggers it. The players are those who tapped a land
// for mana this turn (CR 106.12, the #6217 turn record), read as the trigger
// resolves; each chooses a land of their own, in APNAP order, and every
// chosen land is sacrificed together (CR 101.4, 608.2e). Then each player
// whose sacrificed land was a Plains, as it last existed (CR 608.2h), is
// dealt 2 damage. A player with no land sacrifices nothing and takes none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a80c3af8-470e-4e95-9579-cfdf18fbdbfa",
		Name:         "Desolation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.OldB786EachEndStepBegan,
				"Desolation — each player who tapped a land for mana this turn sacrifices a land; 2 damage to each who sacrificed a Plains",
				effects.OldB786DesolationSacrifice),
		},
	})
}
