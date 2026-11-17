package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Melira, the Living Cure — Legendary Creature — Human Scout {G}{W}, 3/3
// (ONE):
//
//	"If you would get one or more poison counters, instead you get one
//	 poison counter and you can't get additional poison counters this
//	 turn.
//	 Exile Melira: Choose another target creature or artifact. When it's
//	 put into a graveyard this turn, return that card to the battlefield
//	 under its owner's control."
//
// The first half is a CR 614 replacement on the counter window for the
// controller's own poison counters (damage results from infect and toxic
// land through the same window). It sets the amount to one and writes a
// ModCantGetPoisonCounters record for the turn (game/cant_get_poison.go),
// which every later placement asks before the window opens (CR 101.2), so
// the record outlives Melira.
//
// The activated ability exiles Melira as a cost, names another creature or
// artifact as its target, and creates a delayed trigger (CR 603.7) about
// that object: pinned to it for the turn, it fires when that object is put
// into a graveyard from the battlefield, and returns the card to the
// battlefield under its owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bb8c22c-8b01-45f8-8c9e-f683e6b4f022",
		Name:         "Melira, the Living Cure",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCounter && ev.CounterName == game.CounterPoison &&
					ev.CounterPlayer == src.Controller && ev.CounterDelta > 0
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
				ev.CounterDelta = 1
				g.PlayerCantGetPoisonThisTurnForEffect(src.InstanceID, ev.CounterPlayer,
					"Melira, the Living Cure: you can't get additional poison counters this turn")
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Melira, the Living Cure: you get one poison counter instead",
		}},
		Activated: []ActivatedAbility{{
			Label:   "Exile Melira: Choose another target creature or artifact. When it's put into a graveyard this turn, return that card to the battlefield under its owner's control.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ExileThis(),
			Targets: Another(TargetPermanent("another target creature or artifact", Or(Creature(), Artifact()))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				ts := ctx.LegalTargets()
				if len(ts) == 0 {
					return nil
				}
				id := ts[0].ID
				d := g.PinnedTo(g.UntilEndOfTurnDuration(), id)
				g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
					Controller:   ctx.Controller(),
					SourceCardID: ctx.Source(),
					Label:        "Melira, the Living Cure — return that card to the battlefield under its owner's control",
					On:           []game.EventKind{game.EventLTB},
					Condition:    theListedObjectDiedCondition,
					Cards:        []uuid.UUID{id},
					Duration:     &d,
					Body:         meliraReturnListedBody,
				})
				return nil
			},
		}},
	})
}
