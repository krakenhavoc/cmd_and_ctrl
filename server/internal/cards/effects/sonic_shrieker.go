package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sonic Shrieker — Creature — Dragon {2}{R}{W}{B}, 4/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 When this creature enters, it deals 2 damage to any target and you
//	 gain 2 life. If a player is dealt damage this way, they discard a
//	 card."
//
// Flying is a printed keyword. The damage carries a per-recipient
// continuation (Drain Life's recipientTail), told what the recipient was
// actually dealt once the damage has settled: a player dealt more than 0
// discards a card of their choice, so damage that was prevented makes
// nobody discard, and a creature or planeswalker recipient makes nobody
// discard. The life gain is the next step, and it does not depend on the
// damage. With its only target gone the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d317e3f1-553f-44f6-8543-d6d4c42534ad",
		Name:            "Sonic Shrieker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Sonic Shrieker — 2 damage to any target and you gain 2 life; a player dealt damage discards a card",
					Do(Instruction(func(ctx *Context) error {
						t, ok := firstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return DealDamageThen(ctx, t.ID, 2, ctx.Tail(sonicShriekerDiscard))
					}), GainLife{Amount: 2})),
				TargetAny()),
		},
	})
}

// sonicShriekerDiscard is "If a player is dealt damage this way, they
// discard a card", told the recipient and what it was dealt. An on-disk
// key: never renamed, never reused.
var sonicShriekerDiscard = recipientTail("recipient/sonic-shrieker-player-discards", func(ctx *Context, target uuid.UUID, dealt int, _ game.TailOutcome) error {
	if dealt <= 0 || ctx.Game.PlayerByIDForEffect(target) == nil {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: target, Source: ctx.Source(), N: 1})
	return nil
})
