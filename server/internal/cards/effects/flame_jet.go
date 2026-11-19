package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "8a9d6a30-e2c2-4da5-ad3b-d64fb2b5543c",
		Name:         "Flame Jet",
		Completeness: CompletenessFull,
		Targets:      targetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(3),
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
