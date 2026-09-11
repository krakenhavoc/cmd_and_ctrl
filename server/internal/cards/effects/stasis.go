package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stasis — "Players skip their untap steps. At the beginning of
// the upkeep of Stasis's controller, they sacrifice Stasis unless
// they pay {U}."
//
// Both halves are implemented: the skip-untap replacement (S17) and
// the upkeep sacrifice-unless-you-pay-{U} trigger.
//
// #338 stale-simplification sweep: the trigger was left declared
// "lands with S19, sandbox players can manually sacrifice Stasis on
// their upkeep" long after S19 shipped triggered abilities and the
// PayUnless prompt. That left the card with no off switch — a
// Stasis on the battlefield locked the table's untap steps
// indefinitely, because the clause that is supposed to end it every
// upkeep simply never fired.
//
// The replacement AppliesTo fires on any RepEventStepTransition
// where the step being entered is StepUntap, regardless of seat —
// the card's text is "players skip their untap steps", not just
// the controller's. The apply-loop's short-circuit in
// runStepEntryHooksLocked (sub-PR 2) treats any cancel as "skip
// this step" so the turn cursor advances to Upkeep.
func init() {
	Register(Spec{
		OracleID:     "a8cf1379-0195-4e11-b994-481ef1284245",
		Name:         "Stasis",
		Completeness: CompletenessFull,
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
		Triggered: []game.TriggeredAbility{{
			// "At the beginning of the upkeep of Stasis's
			// controller" — the controller's upkeep only, unlike the
			// skip-untap half, which hits every seat.
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Stasis — sacrifice unless you pay {U}",
					func(g *game.Game, item *game.StackItem) error {
						sourceID := item.SourceCardID
						return PayUnless{
							Chooser:  item.Controller,
							Cost:     "{U}",
							Question: "Stasis — pay {U} or sacrifice Stasis?",
							OnDecline: func(ctx *Context) error {
								return SacrificePermanent{Target: sourceID}.Apply(ctx)
							},
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
