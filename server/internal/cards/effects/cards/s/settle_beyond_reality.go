package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Settle Beyond Reality — Sorcery {4}{W} (Jumpstart 2022, #5948):
//
//	"Choose one or both —
//	 • Exile target creature you don't control.
//	 • Exile target creature you control, then return it to the battlefield
//	   under its owner's control."
//
// Go Ninja Go's "choose one or both" with a Flicker bullet (no Controller,
// so it returns under its owner's control), and an exile bullet. The
// bullets run in printed order (CR 608.2c), each on its own target.
//
// Both bullets are shared ModeDoing bodies: Murdock's Crusade's exile of the
// bullet's own target, and Go Ninja Go's flicker of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2debd9f1-6d3a-4f0a-8488-603447c44490",
		Name:         "Settle Beyond Reality",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeDoing("Exile target creature you don't control.",
				effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				effects.MurdocksExileTheModesTarget),
			effects.ModeDoing("Exile target creature you control, then return it to the battlefield under its owner's control.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.FlickerTheModesTarget),
		),
	})
}
