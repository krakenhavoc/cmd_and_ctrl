package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thing from the Deep — Creature — Leviathan {6}{U}{U}{U}, 9/9 (Portal, S82 OLD
// batch 784, #5630):
//
//	"Whenever this creature attacks, sacrifice it unless you sacrifice an
//	 Island."
//
// Scythe Tiger's tax over Islands on an attack trigger: the controller
// sacrifices an Island of their choice, or the Leviathan; with no Island to
// sacrifice it goes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9542bd7f-bb99-4b59-a4e3-b88ed9f798bf",
		Name:         "Thing from the Deep",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Thing from the Deep — sacrifice it unless you sacrifice an Island",
				effects.SacrificeThisUnlessYouSacrificeA("Thing from the Deep", "an Island", effects.Subtype("Island"), effects.S82OldB784AnswerSacrificeAnIslandOrThis)),
		},
	})
}
