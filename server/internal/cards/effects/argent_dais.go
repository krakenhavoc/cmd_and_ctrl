package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Argent Dais — Artifact {1}{W} (Modern Horizons 3, #5948):
//
//	"This artifact enters with two oil counters on it.
//	 Whenever two or more creatures attack, put an oil counter on this
//	 artifact.
//	 {2}, {T}, Remove two oil counters from this artifact: Exile another
//	 target nonland permanent. Its controller draws two cards."
//
// The counters go on through one entry replacement (b10EntersWithCounters). The
// attack trigger is any player's: EventAttack is emitted once per attacking
// creature at the lock-in of one declaration, so the trigger fires on the event
// that brings that player's attackers to two and OncePerBatch declines the rest
// of the declaration, which is one trigger per declaration (CR 603.2c). The
// ability's cost removes two counters as it is announced (an artifact with
// fewer can't activate it); "its controller draws two cards" is the exile's
// tail, run once the move has settled whatever a replacement did, for the
// controller the permanent had as it resolved.
//
// No simplification.
func init() {
	const attackLabel = "Argent Dais — put an oil counter on it"
	Register(Spec{
		OracleID:     "3269f589-00fc-4f0b-8693-656f2ab0670d",
		Name:         "Argent Dais",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters("oil", 2, "Argent Dais: enters with two oil counters"),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack,
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil && len(b13AttackingCreaturesYouControl(g, ev.Actor)) >= 2
				}, attackLabel, Do(CounterOnThis{Kind: "oil", N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Remove two oil counters from this artifact: Exile another target nonland permanent. Its controller draws two cards.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{2}"), TapCost(), RemoveCountersFromThis("oil", 2)),
			Targets: Another(TargetPermanent("another target nonland permanent", Nonland())),
			Effect:  argentDaisExile,
		}},
	})
}

func argentDaisExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		return ExileTarget{
			Target:   t.ID,
			ThenTail: game.KeyedTail{Ref: thenDrawCards, Player: c.Controller, Source: item.SourceCardID, N: 2},
		}.Apply(ctx)
	}
	return nil
}
