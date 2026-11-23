package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adaptive Training Post — Artifact {2}{U} (Tarkir: Dragonstorm Commander):
//
//	"Whenever you cast an instant or sorcery spell, if this artifact has fewer
//	 than three charge counters on it, put a charge counter on it.
//	 Remove three charge counters from this artifact: When you next cast an
//	 instant or sorcery spell this turn, copy it and you may choose new targets
//	 for the copy."
//
// The first ability's "if" is intervening (CR 603.4): checked as the cast
// triggers it and again as it resolves, so a Post that gained its third counter
// from an earlier trigger in the same batch puts no fourth. The second is
// Doublecast's delayed trigger (WhenYouNextCast over the spell cast after the
// ability resolves; the spell that was already cast is not it), with the three
// counters paid at announce as the cost (CR 602.2b), so a response can't spend
// them twice. The copy is made by a triggered ability above the spell, resolves
// first, is not cast, and copies a spell with no targets too (the rulings).
//
// No simplification.
func init() {
	fewerThanThree := func(g *game.Game, source *game.Card) bool {
		return effects.CounterCountOn(g, source.InstanceID, "charge") < 3
	}
	effects.Register(effects.Spec{
		OracleID:     "a6657fcf-f08c-4b03-8ec8-cb0b194eb553",
		Name:         "Adaptive Training Post",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AllOf(effects.YouCast(effects.Or(effects.Instant(), effects.Sorcery())),
				func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return fewerThanThree(g, source)
				}),
				"Adaptive Training Post — put a charge counter on it",
				func(g *game.Game, item *game.StackItem) error {
					if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || !fewerThanThree(g, &c) {
						return nil
					}
					return effects.PutChargeCounterOnThis(g, item)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "Remove three charge counters from this artifact: When you next cast an instant or sorcery spell this turn, copy it and you may choose new targets for the copy.",
			Cost:  effects.RemoveCountersFromThis("charge", 3),
			// It answers nothing: it readies a copy of a spell you cast later (ADR 0142).
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.WhenYouNextCast(
					"Adaptive Training Post — copy that spell",
					game.CastFilter{Types: []string{"Instant", "Sorcery"}},
					effects.CopyTheSpellBody,
				).Apply(effects.NewContext(g, item))
			},
		}},
	})
}
