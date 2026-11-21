package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feroz, Ulgrotha's Warden — Legendary Planeswalker — Feroz {3}{G}{W}{U},
// loyalty 5 (Mystery Booster Commander, #5948):
//
//	"+2: Create two 1/1 blue Bird creature tokens with flying.
//	 0: Draw a card. You may put a permanent card with mana value 4 or less from
//	 your hand onto the battlefield.
//	 −7: Create a Feroz's Ban token. (It's a {6} artifact with "Creature spells
//	 cost {2} more to cast.")
//	 Feroz, Ulgrotha's Warden can be your commander."
//
// The Birds are the table's 1/1 blue flying row. The 0 draws first, then asks
// for a permanent card of mana value 4 or less from the hand, "you may", and puts
// it through the ordinary hand door (Stoneforge Mystic's shape). Feroz's Ban is a
// catalog token (gapB06FerozsBanToken) carrying "Creature spells cost {2} more
// to cast" for every player. "Can be your commander" is deck-building, read from
// the oracle text by the deck validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "724bcd6b-4deb-49c0-aa0e-1245bb40dd56",
		Name:         "Feroz, Ulgrotha's Warden",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Create two 1/1 blue Bird creature tokens with flying.",
				Cost:    LoyaltyCost(2),
				Purpose: game.Purpose{Tokens: 2, CreatureTokens: 2},
				Effect:  Do(CreateToken{Template: TokenCard("1/1 blue Bird with flying"), N: 2}),
			},
			{
				Label:   "0: Draw a card. You may put a permanent card with mana value 4 or less from your hand onto the battlefield.",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 1},
				Effect: Do(DrawCards{N: 1}, PutFromHandOntoBattlefield{
					Match:    ManaValueLE(4),
					Optional: true,
					Label:    "Feroz, Ulgrotha's Warden — you may put a permanent card with mana value 4 or less from your hand onto the battlefield",
				}),
			},
			{
				Label:   "−7: Create a Feroz's Ban token.",
				Cost:    LoyaltyCost(-7),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  Do(CreateToken{Template: gapB06FerozsBanToken(), N: 1}),
			},
		},
	})
}
