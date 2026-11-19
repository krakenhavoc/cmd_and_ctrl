package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worzel, the Protector — Legendary Planeswalker — Worzel {1}{W}{W}{W},
// loyalty 4 (Mystery Booster Commander, #5948):
//
//	"0: Create a 1/1 white Cat creature token with "{T}: Put a loyalty counter
//	 on each planeswalker you control."
//	 −1: Look at the top six cards of your library. You may reveal a
//	 planeswalker or basic Plains card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order.
//	 −8: Create ten Scryb Sprites tokens. (They're {G} 1/1 Faerie creatures with
//	 flying.)
//	 Worzel, the Protector can be your commander."
//
// The Cat is a catalog token (gapB06LoyaltyCatToken) whose tap ability puts a
// loyalty counter on each planeswalker its controller controls as it resolves.
// The −1 is Horn of the Mark's dig (LookAtTopThenMayTakeToHand): a private
// look at six, one planeswalker or basic Plains card revealed into the hand, the
// rest on the bottom at random. The Scryb Sprites are a plain table row, a 1/1
// green Faerie token with flying and the card's own name. "Can be your commander"
// is deck-building, read from the oracle text by the deck validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15084803-8470-469b-8cd9-154c2341b96d",
		Name:         "Worzel, the Protector",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "0: Create a 1/1 white Cat creature token with \"{T}: Put a loyalty counter on each planeswalker you control.\"",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  Do(CreateToken{Template: gapB06LoyaltyCatToken(), N: 1}),
			},
			{
				Label:   "−1: Look at the top six cards of your library. You may reveal a planeswalker or basic Plains card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
				Cost:    LoyaltyCost(-1),
				Purpose: game.Purpose{Tutors: 1},
				Effect: LookAtTopThenMayTakeToHand(6, func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return c.IsPlaneswalker() || (IsBasicLand(c) && c.HasSubtypeWord("Plains"))
				}, 1, "Worzel, the Protector — reveal a planeswalker or basic Plains card from among the top six and put it into your hand"),
			},
			{
				Label:   "−8: Create ten Scryb Sprites tokens.",
				Cost:    LoyaltyCost(-8),
				Purpose: game.Purpose{Tokens: 10, CreatureTokens: 10},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 green Faerie with flying named Scryb Sprites"), N: 10}),
			},
		},
	})
}
