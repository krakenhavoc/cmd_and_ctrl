package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Molten Birth — Sorcery {1}{R}{R} (Magic 2014, S82 OLD batch 401, #5563):
//
//	"Create two 1/1 red Elemental creature tokens, then flip a coin. If you win
//	 the flip, return Molten Birth to its owner's hand."
//
// The tokens are made first, then the keyed coin flip (Boompile's shape): the
// player calls it while the spell is still resolving, and only a won flip marks
// the resolving spell so that its card goes to its owner's hand instead of its
// graveyard as it finishes (game.ReturnResolvingSpellToHandForEffect, the
// resolution exit Soulfire Grand Master uses). A lost flip leaves it to go to
// the graveyard as any sorcery does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aeaae9cc-f30f-4d7e-86f3-3132ef47e987",
		Name:         "Molten Birth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.CreateToken{Template: effects.TokenCard("1/1 red Elemental"), N: 2}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Molten Birth — call the coin flip",
				ThenKey:  moltenBirthFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// moltenBirthFlipped is "if you win the flip, return Molten Birth to its
// owner's hand". An on-disk key.
var moltenBirthFlipped = effects.CoinFlipThen("coin-flip/molten-birth", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) {
		return nil
	}
	ctx.Game.ReturnResolvingSpellToHandForEffect(r.Source)
	return nil
})
