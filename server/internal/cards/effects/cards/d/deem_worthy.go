package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deem Worthy — Instant {4}{R} (Jumpstart 2022, #5948):
//
//	"Deem Worthy deals 7 damage to target creature.
//	 Cycling {3}{R}
//	 When you cycle this card, you may have it deal 2 damage to target
//	 creature."
//
// Complicate's shape: the spell is a targeted damage instant, cycling is the
// keyword's hand ability, and the cycle trigger is an optional targeted
// trigger that works from the graveyard (the card has been discarded as the
// cost, CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "008c92bc-3022-4d34-aa04-584576ff8f8e",
		Name:         "Deem Worthy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 7)),
		OnResolve:    effects.DamageToFirstTargetStep(7),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{3}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Deem Worthy — cycled: you may have it deal 2 damage to target creature",
						effects.GapB171DamageFirstTarget(2))),
					"Deem Worthy — deal 2 damage to target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
