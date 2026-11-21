package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Maiden — Artifact {3}:
//
//	"At the beginning of each opponent's upkeep, this artifact deals X
//	 damage to that player, where X is the number of cards in their
//	 hand minus 4."
//
// Storm World's shape turned around: the trigger fires only on an
// opponent's upkeep, and the damage grows with the hand instead of
// shrinking. X is counted as the trigger resolves (CR 608.2h), and a
// hand of four or fewer takes nothing (CR 107.1b). It is damage from
// the artifact, so prevention and damage triggers see a source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0351638-9224-4779-8f7c-f2b11d134a5f",
		Name:         "Iron Maiden",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Iron Maiden — deals damage to that player equal to their hand size minus 4",
				effects.DamageUpkeepPlayerByHand(effects.HandMinusFour)),
		},
	})
}
