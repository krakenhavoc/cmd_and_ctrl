package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nova Hellkite — Creature — Dragon {3}{R}{R}, 4/5 (Edge of Eternities,
// #4461):
//
//	"Flying, haste
//	 When this creature enters, it deals 1 damage to target creature an
//	 opponent controls.
//	 Warp {2}{R} (You may cast this card from your hand for its warp cost.
//	 Exile this creature at the beginning of the next end step, then you
//	 may cast it from exile on a later turn.)"
//
// Ghitu Slinger's entry damage aimed at a creature an opponent controls,
// with the shared warp cost. Haste is what makes the warp line work: a
// warped Hellkite pings, attacks the turn it arrives, and is exiled at
// the end step to be cast again on a later turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9635d79e-62ec-4d80-ae37-2793489cc13b",
		Name:             "Nova Hellkite",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Warp("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Nova Hellkite — 1 damage to target creature an opponent controls", effects.SourceDealsDamageToEachLegalTarget(1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
		},
	})
}
