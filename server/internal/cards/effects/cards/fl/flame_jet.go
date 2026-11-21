package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame Jet — Sorcery {1}{R} (Urza's Destiny, #5621):
//
//	"Flame Jet deals 3 damage to target player or planeswalker.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Lava Axe's clause at 3 damage, with the engine's cycling row
// (CR 702.29a), which works only while the card is in its owner's
// hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a9d6a30-e2c2-4da5-ad3b-d64fb2b5543c",
		Name:         "Flame Jet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
