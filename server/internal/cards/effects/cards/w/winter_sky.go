package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winter Sky — Sorcery {R} (Homelands, S82 OLD b826, #5635):
//
//	"Flip a coin. If you win the flip, Winter Sky deals 1 damage to each
//	 creature and each player. If you lose the flip, each player draws a
//	 card."
//
// The keyed coin flip (Molten Birth's shape): the caster calls it while the
// spell is resolving. A won flip is Steam Blast's one-instance damage to every
// creature and every player; a lost one draws a card for each player, the
// active player first and then in turn order (CR 101.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fb150d6a-5f81-4fd8-92b2-22a9615bfd4f",
		Name:         "Winter Sky",
		Completeness: effects.CompletenessFull,
		// The damage lands only when the flip is won.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Winter Sky — call the coin flip",
				ThenKey:  winterSkyFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// winterSkyFlipped is the two-way rider on the flip. An on-disk key.
var winterSkyFlipped = effects.CoinFlipThen("coin-flip/s82-old-b826-winter-sky", func(ctx *effects.Context, r game.CoinFlipped) error {
	if effects.WonFirstFlip(r) {
		return effects.DamageEachMatchingAndEachPlayerThen(ctx, effects.Creature(), 1, effects.RecipientThen{})
	}
	for _, id := range effects.GapB34TableInTurnOrder(ctx) {
		if err := (effects.DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
