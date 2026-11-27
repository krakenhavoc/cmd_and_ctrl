package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pale Moon — Instant {1}{U} (Nemesis, S82 OLD batch 679, #5618):
//
//	"Until end of turn, if a player taps a nonbasic land for mana, it produces
//	 colorless mana instead of any other type."
//
// A replacement on the mana production event (CR 614.1a, 106.12b) until end
// of turn, for EVERY player's nonbasic lands (game.ModNonbasicLandManaColorless):
// each mana a nonbasic land makes when tapped becomes {C}, the amount kept.
// Basic lands are untouched, and so is mana that does not come from tapping.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78e3a3c3-3091-46ab-a47a-86b7db68ced4",
		Name:         "Pale Moon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRestrict},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.NonbasicLandsProduceColorlessThisTurnForEffect(item.SourceCardID, item.Controller,
				"Pale Moon — nonbasic lands tapped for mana produce colorless mana")
			return nil
		},
	})
}
