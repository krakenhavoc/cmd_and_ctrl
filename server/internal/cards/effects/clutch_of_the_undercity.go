package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clutch of the Undercity — Instant {1}{U}{U}{B} (Ravnica: City of
// Guilds, #5600):
//
//	"Return target permanent to its owner's hand. Its controller loses 3
//	 life.
//	 Transmute {1}{U}{B} ({1}{U}{B}, Discard this card: Search your
//	 library for a card with the same mana value as this card, reveal it,
//	 put it into your hand, then shuffle. Transmute only as a sorcery.)"
//
// "Its controller" is the player who controlled the permanent, read
// before it moves: a card in a hand has no controller. The life loss is
// the return's continuation (Chain of Vapor's shape, #993), so a bounced
// commander's owner answers the command-zone question first; the loss is
// about a player, not gated on the card reaching the hand, so it happens
// either way. A target gone in response is a fizzle (CR 608.2b) and nobody
// loses life. Transmute is the keyword's constructor (transmute.go, CR
// 702.53).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "316cae8b-31b0-4b76-8b72-8122005e1b4e",
		Name:         "Clutch of the Undercity",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			controller, found := controllerOfTarget(ctx, t.ID)
			if !found {
				return nil
			}
			return BounceToHand{
				Target:   t.ID,
				ThenTail: withCarry(ctx.Tail(clutchOfTheUndercityLoses), controller),
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Transmute("{1}{U}{B}")},
	})
}

// clutchOfTheUndercityLoses is "Its controller loses 3 life", run once the
// return has settled, with that controller as the tail's Carry. An on-disk
// key: never renamed, never reused.
var clutchOfTheUndercityLoses = CardTail("then/clutch-of-the-undercity-controller-loses-3", func(ctx *Context, o game.TailOutcome) error {
	p := tailCarry0(o)
	if p == uuid.Nil {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), p, -3)
})
