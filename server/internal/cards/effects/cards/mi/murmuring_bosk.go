package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murmuring Bosk — Land — Forest:
//
//	"({T}: Add {G}.)
//	 As this land enters, you may reveal a Treefolk card from your hand. If
//	 you don't, this land enters tapped.
//	 {T}: Add {W} or {B}. This land deals 1 damage to you."
//
// The Forest's {G} is the intrinsic mana ability of its subtype (CR 305.6);
// the reveal-land entry over Treefolk cards (a changeling card counts) and
// the painland's coloured half are the card's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42b9d383-3fe2-4fc8-ab86-f80a288d502b",
		Name:         "Murmuring Bosk",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnlessYouRevealFromHand("Murmuring Bosk", "a Treefolk card", func(c game.Card) bool { return c.HasSubtypeWord("Treefolk") }),
		},
		ManaAbilities: []effects.ManaAbility{effects.PainDual("W", "B", "This land")},
	})
}
