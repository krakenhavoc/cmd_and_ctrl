package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Descend upon the Sinful — Sorcery {4}{W}{W} (Outlaws of Thunder Junction
// Commander):
//
//	"Exile all creatures.
//	 Delirium — Create a 4/4 white Angel creature token with flying if there
//	 are four or more card types among cards in your graveyard."
//
// Final Judgment's exile, then the delirium check as the next instruction
// (CR 608.2c), so the sweep is finished first; exiled creatures do not touch
// the graveyard. Each sentence waits for any prompt the one before asks.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cbf95f4-3510-420b-9355-645450f4493d",
		Name:         "Descend upon the Sinful",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepExile}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ExileAllMatching{Match: Creature()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if !drcDeliriumMet(ctx.Game, ctx.Controller()) {
					return nil
				}
				return CreateToken{Template: TokenCard("4/4 white Angel with flying"), N: 1}.Apply(ctx)
			}),
		),
	})
}
