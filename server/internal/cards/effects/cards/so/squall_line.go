package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squall Line — Instant {X}{G}{G} (Time Spiral, #5596):
//
//	"Squall Line deals X damage to each creature with flying and each
//	 player."
//
// Earthquake's mirror image: the shared each-creature-matching-and-each-
// player body at X, over creatures with flying, as one damage event
// (CR 120.3). X = 0 deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "cdf9fab2-bfdb-4bf3-a06a-b21e07600e23",
		Name:     "Squall Line",
		// ADR 0126 §6: only creatures with flying.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.HasKeyword("flying"), x)(ctx.Game, item)
		},
	})
}
