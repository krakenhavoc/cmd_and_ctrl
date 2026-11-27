package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ley Line — Enchantment {3}{G} (Mercadian Masques, S82 OLD batch 698, #5619):
//
//	"At the beginning of each player's upkeep, that player may put a +1/+1
//	 counter on target creature of their choice."
//
// Oath of Druids' shape: the Ley Line's controller controls the trigger, but
// the upkeep player chooses its target (OfThatPlayersChoice) and is the one
// asked whether to put the counter (S82OldB698LeyLineAsk). With no creature on
// the battlefield the trigger is removed (CR 603.3d); a target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "230d79dd-15dd-431d-af22-bdc6eaf86341",
		Name:         "Ley Line",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep("Ley Line — that player may put a +1/+1 counter on target creature of their choice", effects.S82OldB698LeyLineAsk),
			effects.OfThatPlayersChoice(effects.TargetCreature("target creature of their choice")),
		)},
	})
}
