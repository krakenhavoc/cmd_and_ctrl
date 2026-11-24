package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The First Sliver — Legendary Creature — Sliver {W}{U}{B}{R}{G}, 7/7
// (Secret Lair Drop):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)
//	 Sliver spells you cast have cascade."
//
// Imoti's two shapes: it HAS cascade (a trigger on the spell as it is cast) and,
// once on the battlefield, GIVES it to each Sliver spell its controller casts. A
// Sliver spell is judged on the stack, by its types as it was cast, so a
// changeling creature spell counts too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55e4d4b2-1bd8-4c29-a21a-c014a974b715",
		Name:         "The First Sliver",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cascade(),
			effects.GrantsCascade("The First Sliver — Sliver spells you cast have cascade", func(spell game.Card, _ *game.Card, _ *game.Game) bool {
				return spell.HasSubtypeWord("Sliver")
			}),
		},
	})
}
