package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Kaboomist — Creature — Goblin Warrior {1}{R}, 1/2 (Magic 2015,
// #5556):
//
//	"At the beginning of your upkeep, create a colorless artifact token named
//	 Land Mine with '{R}, Sacrifice this token: This token deals 2 damage to
//	 target attacking creature without flying.' Then flip a coin. If you lose
//	 the flip, this creature deals 2 damage to itself."
//
// The Land Mine is made first, then the coin is flipped (the controller calls
// it, as Mogg Assassin's). Losing the flip has the Kaboomist deal 2 damage to
// itself, which a 1/2 does not survive; one that has already left does
// nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5209e1da-aeed-45b3-9640-c2db121d6612",
		Name:         "Goblin Kaboomist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Goblin Kaboomist — create a Land Mine token, then flip a coin; if you lose the flip it deals 2 damage to itself",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CreateToken{Template: effects.OldB356LandMineToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
					g.FlipCoinForEffect(game.CoinFlipSpec{
						Flipper:  item.Controller,
						Source:   item.SourceCardID,
						Question: "Goblin Kaboomist — call the coin flip",
						ThenKey:  goblinKaboomistFlipped,
						Asked:    game.NewAskedItem(item),
					})
					return nil
				}),
		},
	})
}

// goblinKaboomistFlipped is "if you lose the flip, this creature deals 2
// damage to itself". An on-disk key: never renamed, never reused.
var goblinKaboomistFlipped = effects.CoinFlipThen("coin-flip/goblin-kaboomist", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: ctx.Source(), Amount: 2}.Apply(ctx)
})
