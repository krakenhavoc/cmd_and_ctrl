package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nakaya Shade — Creature — Shade {1}{B}, 1/1:
//
//	"{B}: This creature gets +1/+1 until end of turn unless any player
//	 pays {2}."
//
// ADR 0150 (#1903): each activation offers the {2} to every player in
// APNAP order from the active player (AnyPlayerPays); the pump lands
// only if nobody pays, and only on the Shade it was activated from.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce969713-fdaa-4c31-b164-76a9b0b9ff81",
		Name:         "Nakaya Shade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn unless any player pays {2}.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.AnyPlayerPays{
					Cost:     "{2}",
					Question: "Nakaya Shade — pay {2} to stop the +1/+1?",
					Then:     nakayaShadePump,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// nakayaShadePump is "if no one does", keyed (#4264): the asking item carries the Shade's object (CR 400.7).
var nakayaShadePump = effects.AnyPlayerConsequences("any-player/nakaya-shade-pump",
	nil,
	func(ctx *effects.Context, _ game.AnyPlayerAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.ThisGetsUntilEndOfTurn(1, 1, "Nakaya Shade — +1/+1 until end of turn")(ctx.Game, ctx.Item)
	},
	nil)
