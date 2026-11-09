package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eon Fog — Plane — Equilor (Planechase Anthology, #4870):
//
//	"Players skip their untap steps.
//	 Whenever chaos ensues, untap all permanents you control."
//
// A plane (ADR 0216 §6). The first ability is Stasis' replacement of the
// untap step's turn-based action (CR 614.10), read from the command zone
// while the plane is face up. The chaos ability untaps the planar
// controller's permanents (CR 701.26).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a04d05ed-1b9b-4407-8780-c75c102a94dd",
		Name:         "The Eon Fog",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventStepTransition},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventStepTransition && ev.StepTransitionStep == game.StepUntap
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PureCancel: true,
			Label:      "The Eon Fog: players skip their untap steps",
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("The Eon Fog — untap all permanents you control",
				func(g *game.Game, item *game.StackItem) error {
					return b16UntapAllYouControlMatching(NewContext(g, item), item.Controller, func(game.Card) bool { return true })
				}),
		},
	})
}
