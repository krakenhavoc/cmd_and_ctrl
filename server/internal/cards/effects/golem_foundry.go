package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golem Foundry — Artifact {3} (Foundations Commander):
//
//	"Whenever you cast an artifact spell, you may put a charge counter on
//	 this artifact.
//	 Remove three charge counters from this artifact: Create a 3/3
//	 colorless Golem artifact creature token."
//
// A "you may" cast trigger and a counter-removal cost (Magistrate's
// Scepter's RemoveCountersFromThis), paid at announce so a response
// cannot spend the counters twice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e96c09f-c064-4dd8-85ee-0f48d86ab670",
		Name:         "Golem Foundry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && spell.IsArtifact()
			}, "Golem Foundry — you may put a charge counter on this artifact", b33PutChargeCounterOnSelf),
				"Golem Foundry — put a charge counter on this artifact?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Remove three charge counters from this artifact: Create a 3/3 colorless Golem artifact creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
			Cost:    RemoveCountersFromThis(game.CounterCharge, 3),
			Effect:  Do(CreateToken{Template: TokenCard("3/3 colorless Golem artifact creature"), N: 1}),
		}},
	})
}
