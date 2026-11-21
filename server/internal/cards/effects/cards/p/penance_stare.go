package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Penance Stare — Instant {W}:
//
//	"Destroy target attacking creature with mana value 3 or less."
//
// Doom Blade's shape with a combat-state and a mana-value filter on the
// target. Both are read from the creature's effective characteristics,
// when the spell is announced and again as it resolves (CR 608.2b), so a
// creature that stops attacking in response is not destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e21a1ad5-bf31-425f-92c2-7818bdb1cef8",
		Name:         "Penance Stare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target attacking creature with mana value 3 or less",
			effects.And(effects.AttackingCreature(), effects.ManaValueLE(3))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
