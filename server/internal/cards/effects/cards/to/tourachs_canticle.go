package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tourach's Canticle — Sorcery {3}{B}:
//
//	"Target opponent reveals their hand. You choose a card from it.
//	 That player discards that card, then discards a card at random."
//
// The revealed-hand pick (ADR 0116) with a clause that must wait for it
// (#2115, ADR 0116's 2026-10-05 amendment): the random discard is made
// from the hand the chosen card has already left, so it runs in a keyed
// continuation after the discard rather than on the next line. Both
// discards go through the one discard path, so madness applies to each
// (its 2021-06-18 ruling). An empty hand is revealed, nothing is chosen
// and nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "738d2be9-c40b-4a9b-95ed-18d659080918",
		Name:         "Tourach's Canticle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{
				Player: effects.TargetedPlayer(ctx),
				Then:   tourachsCanticleAtRandom,
			}.Apply(ctx)
		},
	})
}

// tourachsCanticleAtRandom is "then discards a card at random".
var tourachsCanticleAtRandom = effects.RevealedPickThen("revealed-pick/tourachs-canticle-discard-at-random",
	func(ctx *effects.Context, pick game.RevealedPick) error {
		return ctx.Game.DiscardRandomForEffect(pick.FromPlayer, 1)
	})
