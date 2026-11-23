package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "c00a0be5-d9a9-4263-a87e-7bf2d128ab3a",
		Name:         "Yoshimaru, Beloved Companion",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventCounterPlaced},
			AppliesTo: effects.PlusOneCounterPlacementOnYourCreature,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Yoshimaru, Beloved Companion: +1 +1/+1 counter",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Put a +1/+1 counter on target legendary creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{6}"),
			Targets: effects.TargetCreature("target legendary creature", effects.Legendary()),
			Effect:  effects.RfCreatureFCounterOnTargetCreature,
		}},
	})
}
