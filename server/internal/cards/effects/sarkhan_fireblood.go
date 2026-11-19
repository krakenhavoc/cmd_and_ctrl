package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarkhan, Fireblood — Legendary Planeswalker — Sarkhan {1}{R}{R}, loyalty 3
// (Foundations Jumpstart, #5929):
//
//	"+1: You may discard a card. If you do, draw a card.
//	 +1: Add two mana in any combination of colors. Spend this mana only to
//	 cast Dragon spells.
//	 −7: Create four 5/5 red Dragon creature tokens with flying."
//
// The first +1 is Battlefield Scavenger's rummage: the discard is optional
// and a card is drawn only for a card discarded. The second is a loyalty
// ability, not a mana ability (CR 605.1a), so it goes on the stack; each of
// the two mana is any colour and carries the restriction "cast a spell with
// the Dragon creature type" (CR 106.6). The −7 makes four 5/5 tokens.
//
// No simplification.
func init() {
	const rummage = "Sarkhan, Fireblood — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:        "074cbe14-a0c8-4772-955a-b4052333d6ce",
		Name:            "Sarkhan, Fireblood",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label: "+1: You may discard a card. If you do, draw a card.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(1, false, rummage, drawAsManyAsDiscarded)(NewContext(g, item))
				},
			},
			{
				Label: "+1: Add two mana in any combination of colors. Spend this mana only to cast Dragon spells.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{
						Produced:     AnyCombinationOfColors(2),
						Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dragon")},
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−7: Create four 5/5 red Dragon creature tokens with flying.",
				Cost:    LoyaltyCost(-7),
				Purpose: game.Purpose{Tokens: 4},
				Effect:  Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 4}),
			},
		},
	})
}
