package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stasis — "Players skip their untap steps. At the beginning of
// the upkeep of Stasis's controller, they sacrifice Stasis unless
// they pay {U}."
//
// S17 ships only the skip-untap half via a RepEventStepTransition
// replacement. The upkeep-sacrifice half is a triggered ability
// that lands with S19. Documented simplification: sandbox players
// can manually sacrifice Stasis on their upkeep until S19 wires
// the trigger.
//
// The replacement AppliesTo fires on any RepEventStepTransition
// where the step being entered is StepUntap, regardless of seat —
// the card's text is "players skip their untap steps", not just
// the controller's. The apply-loop's short-circuit in
// runStepEntryHooksLocked (sub-PR 2) treats any cancel as "skip
// this step" so the turn cursor advances to Upkeep.
func init() {
	Register(Spec{
		OracleID: "a8cf1379-0195-4e11-b994-481ef1284245",
		Name:     "Stasis",
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventStepTransition},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventStepTransition && ev.StepTransitionStep == game.StepUntap
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.Cancel()
					return nil
				},
				Controller: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Stasis: skip untap step",
			},
		},
	})
}
