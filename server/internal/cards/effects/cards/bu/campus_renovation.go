package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Campus Renovation — Sorcery {3}{R}{W} (March of the Machine: The
// Aftermath, #5476):
//
//	"Return up to one target artifact or enchantment card from your
//	 graveyard to the battlefield. Exile the top two cards of your library.
//	 Until the end of your next turn, you may play those cards."
//
// The return is "up to one", so the spell is castable with no target and
// the exile still happens. A target that left the graveyard in response
// does nothing (CR 608.2b). The exile is Atsushi's window
// (ExileTopNUntilYourNextTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d34a3b58-a905-4e90-b35c-9f5c21149b04",
		Name:         "Campus Renovation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.MatArtifactOrEnchantmentFromYourGraveyard("up to one target artifact or enchantment card from your graveyard").WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			effects.MatReturnTargetToBattlefield(ctx)
			return effects.ExileTopNUntilYourNextTurn(ctx, 2)
		},
	})
}
