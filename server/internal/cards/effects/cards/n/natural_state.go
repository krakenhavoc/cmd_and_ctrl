package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural State — Instant {G} (Oath of the Gatewatch, #5545):
//
//	"Destroy target artifact or enchantment with mana value 3 or less."
//
// Abrupt Decay's mana value clause on an artifact or enchantment; a card
// with no mana cost has mana value 0 (CR 202.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f500f4cc-d5ee-4995-b7b5-96a5f63dfb47",
		Name:         "Natural State",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target artifact or enchantment with mana value 3 or less",
			effects.And(effects.Or(effects.Artifact(), effects.Enchantment()), effects.ManaValueLE(3))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
