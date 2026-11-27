package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baku Altar — Artifact {2} (Betrayers of Kamigawa, #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki counter on
//	 this artifact.
//	 {2}, {T}, Remove a ki counter from this artifact: Create a 1/1 colorless
//	 Spirit creature token."
//
// Quillmane Baku's trigger on an artifact, and a one-counter removal paid as
// the ability is activated, so a response cannot spend the same counter
// twice. The token is Spiritual Visit's 1/1 colorless Spirit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8443b8c2-d0ab-4fb1-ad35-7e5e4cb345b6",
		Name:         "Baku Altar",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Subtype("Spirit"), effects.Subtype("Arcane")),
				"Baku Altar — put a ki counter on this artifact", effects.PutACounterOnThis("ki")),
				"Baku Altar — put a ki counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Remove a ki counter from this artifact: Create a 1/1 colorless Spirit creature token.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.RemoveCountersFromThis("ki", 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 colorless Spirit"), N: 1}.Apply(ctx)
			},
		}},
	})
}
