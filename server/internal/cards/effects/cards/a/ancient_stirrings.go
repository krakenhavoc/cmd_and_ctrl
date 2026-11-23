package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ancient Stirrings — Sorcery {G} (Mystery Booster 2, #5948):
//
//	"Look at the top five cards of your library. You may reveal a colorless
//	 card from among them and put it into your hand. Then put the rest on
//	 the bottom of your library in any order."
//
// Commune with Nature's dig over colorless cards (lands and artifacts
// count): optional, at most one, revealed, the rest to the bottom in an order
// the player chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82f18e7d-5c42-47c4-8e74-3fccc9b7b1f0",
		Name:         "Ancient Stirrings",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.Colorless(),
				Max:      1,
				Optional: true,
				Label:    "Ancient Stirrings — reveal a colorless card and put it into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
