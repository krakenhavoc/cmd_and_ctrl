package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Intervene — Instant {U}: "Counter target spell that targets a creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "380b386f-90c8-45c5-a611-1ed6803b52d2",
		Name:         "Intervene",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell that targets a creature", effects.OldB637SpellTargetsACreature()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
