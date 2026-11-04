package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relic Retriever — Creature — Spirit Monkey {1}{R}, 2/1 (Secrets of
// Strixhaven Commander, #4834):
//
//	"First strike
//	 At the beginning of each end step, if a card left your graveyard this
//	 turn, create a Treasure token."
//
// Gau, Feral Youth's end-step shape: an intervening "if" (CR 603.4) read
// off the turn's zone moves out of a graveyard of a card its controller
// owns (cardLeftYourGraveyardThisTurn), checked as each end step begins and
// again as the trigger resolves (InterveningIf). It is "each" end step, an
// opponent's too (the 2026-03-20 ruling: it checks as the step starts).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "689c41e9-800a-4998-bb04-74bf90141035",
		Name:            "Relic Retriever",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(cardLeftYourGraveyardThisTurnCond, game.TriggeredAbility{
				Watches: []game.EventKind{game.EventStepBegan},
				AppliesTo: AllOf(
					StepBegan(game.StepEnd, false),
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return cardLeftYourGraveyardThisTurn(g, source.Controller)
					},
				),
				Key:    "Relic Retriever — create a Treasure token",
				Effect: Do(CreateToken{Template: TreasureToken(), N: 1}),
			}),
		},
	})
}

// cardLeftYourGraveyardThisTurnCond is the intervening "if a card left your
// graveyard this turn" (CR 603.4) as a Do step, read off this turn's zone
// moves.
var cardLeftYourGraveyardThisTurnCond = CondBoard{
	Name: "a card left your graveyard this turn",
	Test: func(ctx *Context) bool { return cardLeftYourGraveyardThisTurn(ctx.Game, ctx.Controller()) },
}
