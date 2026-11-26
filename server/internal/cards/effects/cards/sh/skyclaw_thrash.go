package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyclaw Thrash — Artifact Creature — Lizard Warrior {3}{U}{R}, 4/4 (Alara
// Reborn):
//
//	"Whenever this creature attacks, flip a coin. If you win the flip, this
//	 creature gets +1/+1 and gains flying until end of turn."
//
// The called flip (CR 705.2) is Boompile's keyed one; a won flip pumps and
// gives flying to the Thrash only while it is still the permanent that
// attacked (CR 400.7). A lost flip does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d11cdb2-5a44-4e0e-b357-bb1be59f68fb",
		Name:         "Skyclaw Thrash",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Skyclaw Thrash — flip a coin", func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Skyclaw Thrash — call the coin flip",
					ThenKey:  skyclawThrashFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			}),
		},
	})
}

// skyclawThrashFlipped is "if you win the flip, this creature gets +1/+1
// and gains flying until end of turn". An on-disk key: never renamed,
// never reused.
var skyclawThrashFlipped = effects.CoinFlipThen("coin-flip/skyclaw-thrash", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) || !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	id := ctx.Item.SourceCardID
	if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Skyclaw Thrash — +1/+1"}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Skyclaw Thrash — flying"}.Apply(ctx)
})
