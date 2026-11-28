package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Illicit Auction — Sorcery {3}{R}{R} (Mirage, S82 OLD batch 809, #5633):
//
//	"Each player may bid life for control of target creature. You start the
//	 bidding with a bid of 0. In turn order, each player may top the high bid.
//	 The bidding ends if the high bid stands. The high bidder loses life equal
//	 to the high bid and gains control of the creature. (This effect lasts
//	 indefinitely.)"
//
// LifeBidding (#6212) runs the auction over every player. The win is a
// registered number branch: the high bidder loses the bid in life
// (CR 119.3) and gains control of the creature, if it is still a legal
// target (CR 608.2b), with no end (CR 611.2a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dea8c725-455e-4668-877a-df388fe22255",
		Name:         "Illicit Auction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(effects.LifeBidding{
			Opening: 0,
			Then:    illicitAuctionWon,
		}),
	})
}

// illicitAuctionWon is the high bidder's win. An on-disk key: never renamed,
// never reused.
var illicitAuctionWon = effects.NumberThen("number/illicit-auction-won", func(ctx *effects.Context, r game.NumberAnswer) error {
	if r.Chooser == uuid.Nil {
		return nil
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -r.Number); err != nil {
		return err
	}
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.GainControl{
		Target:     t.ID,
		Controller: r.Chooser,
		Duration:   game.IndefiniteDuration(),
		Label:      "Illicit Auction — gain control (no stated duration)",
	}.Apply(ctx)
})
