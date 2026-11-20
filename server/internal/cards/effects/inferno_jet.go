package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inferno Jet — Sorcery {5}{R} (Hour of Devastation, #5534):
//
//	"Inferno Jet deals 6 damage to target opponent or planeswalker.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The target is an opponent (never you) or a planeswalker
// (targetOpponentOrPlaneswalker, Zealot of the God-Pharaoh's), and the
// shared Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50c13b38-9865-4c60-b2df-3fd6e3a9c6a2",
		Name:         "Inferno Jet",
		Completeness: CompletenessFull,
		Targets:      targetOpponentOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 6)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(6),
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
