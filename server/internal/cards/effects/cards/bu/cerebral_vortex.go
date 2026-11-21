package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cerebral Vortex — Instant {1}{U}{R} (Guildpact, #5599):
//
//	"Target player draws two cards, then Cerebral Vortex deals damage to
//	 that player equal to the number of cards they've drawn this turn."
//
// Two steps (ResolveSteps, CR 608.2c): the two draws, then the damage,
// which waits for any question a draw asked. The count is the player's
// turn tally of cards drawn, read once the draws are done, so it
// includes these two and every card they drew earlier this turn, their
// draw step's included. A draw replaced by something else was not drawn
// and does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c68b4003-e539-4b88-acc9-e95c100173c0",
		Name:         "Cerebral Vortex",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.ResolveSteps(effects.Instruction(cerebralVortexDraw), effects.Instruction(cerebralVortexDamage)),
	})
}

// cerebralVortexDraw is "target player draws two cards".
func cerebralVortexDraw(ctx *effects.Context) error {
	p := effects.TargetedPlayer(ctx)
	if p == uuid.Nil {
		return nil
	}
	return effects.DrawCards{Player: p, N: 2}.Apply(ctx)
}

// cerebralVortexDamage deals the targeted player damage equal to the
// cards they have drawn this turn.
func cerebralVortexDamage(ctx *effects.Context) error {
	p := effects.TargetedPlayer(ctx)
	if p == uuid.Nil {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: p, Amount: ctx.Game.TurnTallyFor(p).CardsDrawn}.Apply(ctx)
}
