package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zimone, Quandrix Prodigy — Legendary Creature — Human Wizard {G}{U},
// 1/2 (Secrets of Strixhaven Commander, #4834):
//
//	"{1}, {T}: You may put a land card from your hand onto the
//	 battlefield tapped.
//	 {4}, {T}: Draw a card. If you control eight or more lands, draw two
//	 cards instead."
//
// The first ability is Herbie's put-from-hand: an optional pick among the
// land cards in hand, entering tapped (the entry carries the tapped, so
// nothing sees an untapped land). With no land in hand nothing is asked.
// The second counts the lands the controller controls as it resolves
// (CR 608.2), not as it was activated, and draws two cards instead of one
// when there are eight or more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee2ef1e0-6803-4eb5-8664-605438c67505",
		Name:         "Zimone, Quandrix Prodigy",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: You may put a land card from your hand onto the battlefield tapped.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  Do(MayPutALandFromHandTapped("Zimone, Quandrix Prodigy")),
			},
			{
				Label:   "{4}, {T}: Draw a card. If you control eight or more lands, draw two cards instead.",
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
				Effect: Do(Instruction(func(ctx *Context) error {
					n := 1
					if countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsLand() }) >= 8 {
						n = 2
					}
					return DrawCards{N: n}.Apply(ctx)
				})),
			},
		},
	})
}
