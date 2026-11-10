package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rush of Dread — Sorcery {1}{B}{B} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {1} — Target opponent sacrifices half the creatures they control of
//	   their choice, rounded up.
//	 + {2} — Target opponent discards half the cards in their hand, rounded up.
//	 + {2} — Target opponent loses half their life, rounded up."
//
// Each bullet has its own target opponent (they may be three different
// players). Every count is worked out as that bullet resolves, once, from what
// the opponent has then: the creatures (a creature of their choice, the
// Pox Plague / Zodiark sacrifice seam, PlayersSacrificeForEffect, with the
// count rounded up), the cards in hand (a discard they choose, the count passed
// as N because DiscardPrompt.HalfHand rounds down) and their life total (an
// opponent at 0 or less loses nothing). A target gone by resolution is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	opp := func() *game.TargetSpec { return TargetPlayer("target opponent", Opponent()) }
	Register(Spec{
		OracleID:     "0dfe6ec6-a0df-41c8-90dd-1ddf0417700a",
		Name:         "Rush of Dread",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Target opponent sacrifices half the creatures they control of their choice, rounded up.", "{1}", opp(), rushOfDreadSacrifice),
			SpreeModeDoing("Target opponent discards half the cards in their hand, rounded up.", "{2}", opp(), rushOfDreadDiscard),
			SpreeModeDoing("Target opponent loses half their life, rounded up.", "{2}", opp(), rushOfDreadLoseLife),
		),
	})
}

// rushOfDreadSacrifice is the first bullet: the creatures they control now, half
// of them rounded up, each a creature of their choice.
func rushOfDreadSacrifice(_ *game.StackItem, ctx *Context, occ int) error {
	who, ok := rushOfDreadPlayer(ctx, occ)
	if !ok {
		return nil
	}
	owed := (len(permanentsControlledByMatching(ctx.Game, who, Creature())) + 1) / 2
	if owed == 0 {
		return nil
	}
	asks := make([]uuid.UUID, owed)
	for i := range asks {
		asks[i] = who
	}
	ctx.Game.PlayersSacrificeForEffect(ctx.Source(), asks, sacrificeSpec("a creature", Creature()),
		"Rush of Dread — sacrifice half the creatures you control, rounded up")
	return nil
}

// rushOfDreadDiscard is the second bullet: half the cards in their hand,
// rounded up, discarded as they choose. Nothing in the spell waits on the answer.
func rushOfDreadDiscard(_ *game.StackItem, ctx *Context, occ int) error {
	who, ok := rushOfDreadPlayer(ctx, occ)
	if !ok {
		return nil
	}
	p := ctx.PlayerByID(who)
	if p == nil || p.Hand.Size() == 0 {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   who,
		Source:   ctx.Source(),
		N:        (p.Hand.Size() + 1) / 2,
		Question: "Rush of Dread — discard half the cards in your hand, rounded up",
	})
	return nil
}

// rushOfDreadLoseLife is the third bullet: half their life, rounded up; an
// opponent at 0 or less loses nothing.
func rushOfDreadLoseLife(_ *game.StackItem, ctx *Context, occ int) error {
	who, ok := rushOfDreadPlayer(ctx, occ)
	if !ok {
		return nil
	}
	p := ctx.PlayerByID(who)
	if p == nil || p.Life <= 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), who, -((p.Life + 1) / 2))
}

// rushOfDreadPlayer is the opponent a bullet targeted, if still legal.
func rushOfDreadPlayer(ctx *Context, occ int) (uuid.UUID, bool) {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
		return uuid.Nil, false
	}
	return t.ID, true
}
