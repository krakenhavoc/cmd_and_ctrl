package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secluded Glen — Land (Wilds of Eldraine Commander, #5368):
//
//	"As this land enters, you may reveal a Faerie card from your hand. If you
//	 don't, this land enters tapped.
//	 {T}: Add {U} or {B}."
//
// The reveal-land entry over Faerie cards: the clause is the card's own and
// the test is the type line of a card in hand (no layer applies there), a
// changeling card counting as a Faerie.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09f52275-99e6-45e0-b2db-cafe26d5fb91",
		Name:         "Secluded Glen",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnlessYouRevealFromHand("Secluded Glen", "a Faerie card", effects.Cmdr2S12IsFaerieCard),
		},
		ManaAbilities: []effects.ManaAbility{effects.DualManaAbility("U", "B")},
	})
}
