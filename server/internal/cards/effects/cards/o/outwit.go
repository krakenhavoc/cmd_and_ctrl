package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Outwit — Instant {U}:
//
//	"Counter target spell that targets a player."
//
// A spell that targets only permanents or spells is not a legal target.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eac02e36-15c8-4b6d-b9b4-da549a722336",
		Name:         "Outwit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell that targets a player", effects.SpellTargetsAPlayer()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
