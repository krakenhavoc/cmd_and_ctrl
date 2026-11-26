package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Castigator — Creature — Kor Wizard Ally {1}{W}, 3/1 (Battle for
// Zendikar, #5547):
//
//	"This creature can't be blocked by Eldrazi Scions."
//
// A block rule on the creature itself: a creature that is an Eldrazi Scion
// (both subtypes, read through the layered view) can't block it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbb9dafa-b1aa-401a-b26c-c0c7e8b7222f",
		Name:         "Kor Castigator",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.HasSubtypeWord("Eldrazi") && c.HasSubtypeWord("Scion")
			}, "Eldrazi Scions"),
		},
	})
}
