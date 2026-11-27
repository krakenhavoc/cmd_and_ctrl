package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elephant Resurgence — Sorcery {1}{G} (Prophecy, S82 OLD batch 675, #5617):
//
//	"Each player creates a green Elephant creature token. Those creatures have
//	 "This token's power and toughness are each equal to the number of creature
//	 cards in its controller's graveyard.""
//
// Each player, from the active player in turn order, creates an Elephant of
// their own. The ability is printed on the token (a catalog token template, ADR
// 0083): a layer 7a characteristic-defining static counted live in the token's
// controller's graveyard, so each player's Elephant is as big as their own
// creature cards, whoever cast the spell, and it is 0/0 and dies to the
// state-based check while that graveyard has none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2d54c59-7722-4520-adca-cd02a0871ab2",
		Name:         "Elephant Resurgence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, p := range effects.SeatsFromActive(ctx.Game) {
				if err := (effects.CreateToken{Controller: p, Template: effects.S82OldB675ElephantToken(), N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
