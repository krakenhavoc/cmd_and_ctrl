package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ce969713-fdaa-4c31-b164-76a9b0b9ff81",
		Name:         "Nakaya Shade",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn unless any player pays {2}.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return AnyPlayerPays{
					Cost:     "{2}",
					Question: "Nakaya Shade — pay {2} to stop the +1/+1?",
					OnNobody: func(ctx *Context) error {
						if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						return thisGetsUntilEndOfTurn(1, 1, "Nakaya Shade — +1/+1 until end of turn")(ctx.Game, ctx.Item)
					},
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
