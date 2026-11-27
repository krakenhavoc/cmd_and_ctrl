package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep-Slumber Titan — Creature — Giant Warrior {2}{R}{R}, 7/7 (Shadowmoor, #5591):
//
//	"This creature enters tapped.
//	 This creature doesn't untap during your untap step.
//	 Whenever this creature is dealt damage, untap it."
//
// Enters tapped is Billiard Room's replacement; the untap-step restriction
// is Battered Golem's; the damage trigger is one per damage instance
// (Ripjaw Raptor's enrage shape, CR 120.3) and untaps the Titan only while
// it is still the permanent that was damaged.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "d8f2c12c-d463-42dc-95f1-2daf68f99873",
		Name:                  "Deep-Slumber Titan",
		Completeness:          effects.CompletenessFull,
		Replacements:          []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Deep-Slumber Titan — untap it",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.B31UntapSelf,
		}},
	})
}
