package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Senate Guildmage — Creature — Human Wizard {W}{U}, 2/2 (Ravnica
// Allegiance, #5521):
//
//	"{W}, {T}: You gain 2 life.
//	 {U}, {T}: Draw a card, then discard a card."
//
// Two tap abilities: Ghost-Lit Redeemer's {W}, {T} life gain and Frontline
// Sage's {U}, {T} loot (lootOne, which waits for the discard choice).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76f1cc7a-084e-410d-92cb-b8f6ab8da4e9",
		Name:         "Senate Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, {T}: You gain 2 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{W}"), TapCost()),
				Effect:  Do(GainLife{Amount: 2}),
			},
			{
				Label:   "{U}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
				Cost:    Plus(ManaCost("{U}"), TapCost()),
				Effect:  func(g *game.Game, item *game.StackItem) error { return lootOne(g, item, 1) },
			},
		},
	})
}
