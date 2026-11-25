package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elaborate Firecannon — Artifact {2} (Ixalan, #5532):
//
//	"This artifact doesn't untap during your untap step.
//	 {4}, {T}: This artifact deals 2 damage to any target.
//	 At the beginning of your upkeep, you may discard a card. If you do,
//	 untap this artifact."
//
// The upkeep "may" is the discard itself, chosen up to one card as the
// trigger resolves; the untap happens only if a card was discarded
// (CR 701.9a), and only while this is still the same permanent (CR 400.7).
//
// No simplification.
func init() {
	const label = "Elaborate Firecannon — you may discard a card; if you do, untap this artifact"
	effects.Register(effects.Spec{
		OracleID:              "6e61ae7b-f2e8-421c-870b-a48c08d07503",
		Name:                  "Elaborate Firecannon",
		Completeness:          effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: This artifact deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect:  effects.DamageToFirstTarget(2),
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   item.Controller,
					Source:   item.SourceCardID,
					N:        1,
					UpTo:     true,
					Question: label + "?",
				}, effects.FirecannonUntapIfDiscarded.Then, nil)
			}),
		},
	})
}
