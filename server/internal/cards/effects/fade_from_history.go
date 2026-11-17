package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "533767ca-7208-4c88-b160-375479b68285",
		Name:         "Fade from History",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep:  game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy, Partial: true},
			Tokens: 1,
		},
		OnResolve: ResolveSteps(
			Instruction(fadeFromHistoryBears),
			Instruction(func(ctx *Context) error {
				return DestroyAllMatching{Match: Or(Artifact(), Enchantment())}.Apply(ctx)
			}),
		),
	})
}

// fadeFromHistoryBears is "each player who controls an artifact or
// enchantment creates a 2/2 green Bear creature token."
func fadeFromHistoryBears(ctx *Context) error {
	g := ctx.Game
	var makers []uuid.UUID
	for _, p := range apnapPlayers(g) {
		if b40ControlsA(g, p, Or(Artifact(), Enchantment())) {
			makers = append(makers, p)
		}
	}
	for _, p := range makers {
		if err := (CreateToken{Controller: p, Template: TokenCard("2/2 green Bear"), N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
