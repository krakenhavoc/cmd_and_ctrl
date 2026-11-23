package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cut Short — Instant {2}{W} (March of the Machine):
//
//	"Convoke
//	 Destroy target planeswalker that was activated this turn or tapped
//	 creature."
//
// One target: a planeswalker one of whose loyalty abilities was activated
// this turn (Game.LoyaltyActivatedThisTurn), or a tapped creature. The
// state is checked as the spell is cast and again as it resolves
// (CR 608.2b): a creature that untapped is no longer a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "428d131c-b0fa-4e96-9c02-e2e6714357b9",
		Name:         "Cut Short",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Targets: effects.TargetPermanent("target planeswalker that was activated this turn or tapped creature",
			effects.ActivatedPlaneswalkerOrTapped()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
