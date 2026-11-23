package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fade from History — Sorcery {2}{G}{G} (The Brothers' War, #5482):
//
//	"Each player who controls an artifact or enchantment creates a 2/2
//	 green Bear creature token. Then destroy all artifacts and
//	 enchantments."
//
// Two instructions in printed order (CR 608.2c, ResolveSteps): the first
// reads the board once, before anything is destroyed, so every player who
// controls an artifact or enchantment gets a Bear (in turn order), whatever
// the second then destroys; the Bears are neither, so they stay.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "533767ca-7208-4c88-b160-375479b68285",
		Name:         "Fade from History",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Sweep:  game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy, Partial: true},
			Tokens: 1,
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(fadeFromHistoryBears),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyAllMatching{Match: effects.Or(effects.Artifact(), effects.Enchantment())}.Apply(ctx)
			}),
		),
	})
}

// fadeFromHistoryBears is "each player who controls an artifact or
// enchantment creates a 2/2 green Bear creature token."
func fadeFromHistoryBears(ctx *effects.Context) error {
	g := ctx.Game
	var makers []uuid.UUID
	for _, p := range effects.ApnapPlayers(g) {
		if effects.B40ControlsA(g, p, effects.Or(effects.Artifact(), effects.Enchantment())) {
			makers = append(makers, p)
		}
	}
	for _, p := range makers {
		if err := (effects.CreateToken{Controller: p, Template: effects.TokenCard("2/2 green Bear"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
