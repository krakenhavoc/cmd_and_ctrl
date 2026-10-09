package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yoshimaru, Beloved Companion — Legendary Creature — Dog {2}{W}, 2/2:
//
//	"If one or more +1/+1 counters would be put on a creature you
//	 control, that many plus one +1/+1 counters are put on it instead.
//	 {6}: Put a +1/+1 counter on target legendary creature."
//
// Hardened Scales' replacement (plusOneCounterPlacementOnYourCreature),
// on a body; the activation's counter goes through the same pipeline, so
// it is placed as two on a creature you control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c00a0be5-d9a9-4263-a87e-7bf2d128ab3a",
		Name:         "Yoshimaru, Beloved Companion",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: plusOneCounterPlacementOnYourCreature,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Yoshimaru, Beloved Companion: +1 +1/+1 counter",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Put a +1/+1 counter on target legendary creature.",
			Cost:    ManaCost("{6}"),
			Targets: TargetCreature("target legendary creature", Legendary()),
			Effect:  rfCreatureFCounterOnTargetCreature,
		}},
	})
}
