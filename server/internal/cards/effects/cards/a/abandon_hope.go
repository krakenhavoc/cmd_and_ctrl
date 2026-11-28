package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abandon Hope — {X}{1}{B} Sorcery (Tempest, S82 OLD batch 769, #5628):
//
//	"As an additional cost to cast this spell, discard X cards.
//	 Look at target opponent's hand and choose X cards from it. That player
//	 discards those cards."
//
// Channeled Force's cost (DiscardXCost): the X announced with the cast is also the
// number of cards discarded while the spell is on the stack (CR 601.2f). The
// resolution is Seer's Vision's look-and-choose over X cards: only the caster
// looks (the hand is not revealed), the caster picks X of them, or the whole hand
// when it holds fewer, and the opponent discards those cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "8adbba6e-03ef-4278-aec5-8a4496b377a8",
		Name:           "Abandon Hope",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.DiscardXCost(),
		Targets:        effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.X() <= 0 {
				return nil
			}
			return effects.ChooseFromRevealedHand{Player: effects.TargetedPlayer(ctx), Label: "card", Look: true, Count: ctx.X()}.Apply(ctx)
		},
	})
}
