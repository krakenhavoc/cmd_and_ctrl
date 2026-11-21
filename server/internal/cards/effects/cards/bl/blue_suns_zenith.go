package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blue Sun's Zenith — Instant {X}{U}{U}{U}:
//
//	"Target player draws X cards.
//	 Shuffle Blue Sun's Zenith into its owner's library."
//
// An X spell that never leaves. ctx.X() reads back the announced
// value — the cost engine already charged X·generic at cast time —
// and the second sentence replaces the ordinary "goes to the
// graveyard on resolution" routing with a tuck-and-shuffle: the
// resolving card is still on the stack when OnResolve runs (the
// engine only routes it away AFTER OnResolve returns), so tucking it
// here is exactly the "the spell may have moved itself" case
// resolveTopOfStackLocked already special-cases (CR 608.2n) — it
// notices the card left the stack on its own and skips the graveyard
// route.
//
// "Shuffle into" is put-on-top-then-shuffle, the same pair Teferi
// Akosa's reflexive tuck uses: the card's exact landing spot doesn't
// matter because the very next step randomises it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "613a41b8-0b4f-4995-bf1e-ca41f96e6438",
		Name:         "Blue Sun's Zenith",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			// The tuck waits for a prompt the draws ask (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				if t := ctx.Item.Targets; len(t) == 0 || t[0].Kind != game.TargetPlayer {
					return nil
				}
				return effects.DrawCards{Player: ctx.Item.Targets[0].ID, N: ctx.X()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if t := ctx.Item.Targets; len(t) == 0 || t[0].Kind != game.TargetPlayer {
					return nil
				}
				owner := ctx.Item.Owner
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Source()}, game.TuckOptions{}, game.KeyedTail{Ref: effects.ThenShuffle, Player: owner})
			}),
		),
	})
}
