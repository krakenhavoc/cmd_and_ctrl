package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mire's Toll — Sorcery {B} (Worldwake, S82 OLD b488):
//
//	"Target player reveals a number of cards from their hand equal to the number
//	 of Swamps you control. You choose one of them. That player discards that
//	 card."
//
// The shared reveal-N-and-pick flow (Acquisitions Expert's): the Swamps are counted as the
// spell resolves, through the effective land types ("Swamp" read through the card's text),
// the target player chooses which cards to reveal (a hand no larger than the count is
// revealed whole), only those are revealed, and you choose one of them for that player to
// discard. No Swamp, or an empty hand, reveals and asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e0756d8-c2bd-40f1-a281-0828c5a2da03",
		Name:         "Mire's Toll",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.TargetedPlayer(ctx)
			if target == uuid.Nil {
				return nil
			}
			return miresTollReveal.Ask(ctx, target, effects.SwampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp")))
		},
	})
}

// miresTollReveal is the reveal-N-and-pick flow. Both keys are on-disk identities:
// never renamed, never reused.
var miresTollReveal = effects.NewRevealFromHandPick("Mire's Toll", "discards",
	"choose-cards/mires-toll-reveal", "choose-cards/mires-toll-discard",
	func(ctx *effects.Context, owner, card uuid.UUID) error {
		return ctx.Game.DiscardCardsKeyedForEffect(owner, []uuid.UUID{card}, ctx.Source(), game.KeyedTail{})
	})
