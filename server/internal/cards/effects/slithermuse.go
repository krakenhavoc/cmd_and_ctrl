package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slithermuse — "When this creature leaves the battlefield, choose
// an opponent. If that player has more cards in hand than you, draw
// cards equal to the difference."
//
// A blink deck's draw engine: it triggers on ANY leave, not just
// death, so flickering it refills your hand.
//
// S22: evoke works. "Evoke {3}{U} (You may cast this spell for its
// evoke cost. If you do, it's sacrificed when it enters.)" is the
// deck's actual play pattern — pay one less, let it die on entry, and
// the leaves-the-battlefield trigger above refills your hand at
// instant speed if you have a way to flash it in.
//
// The sacrifice is a triggered ability rather than part of the
// resolution (CR 702.74a), which is what makes the sequence work: the
// creature genuinely enters, the sacrifice goes on the stack where it
// can be responded to, and the leave-trigger then goes on the stack
// above the empty board and draws. Folding the sacrifice into
// resolution would produce the same hand size and the wrong stack.
//
// #929: "choose an opponent" is a real prompt. It used to be
// auto-picked as the opponent holding the most cards — always the
// choice that drew the most, and never the player's. The
// choose-a-player primitive (effects/choose_player.go) is the
// trigger-side player picker that simplification was waiting on, so
// the caveat is gone and the controller picks.
func init() {
	Register(Spec{
		OracleID:     "4b6512aa-535e-4edf-8797-43df2b8463de",
		Name:         "Slithermuse",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{3}{U}"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, Self, "Slithermuse — draw the hand-size difference", func(g *game.Game, item *game.StackItem) error {
				return ChoosePlayer{
					Among:    Opponents,
					Question: "Slithermuse — choose an opponent",
					Then:     slithermuseDraw,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// slithermuseDraw is the "if that player has more cards in hand than
// you, draw cards equal to the difference" half, run once the
// controller has named the opponent.
//
// A package-level function reading only the Context it is handed —
// the StackItem.Effect contract — so an undo across the prompt
// resolves it against the restored game.
func slithermuseDraw(ctx *Context) error {
	opp := ctx.PlayerByID(ctx.ChosenPlayer())
	me := ctx.PlayerByID(ctx.Controller())
	if opp == nil || me == nil {
		return nil
	}
	diff := len(opp.Hand.Cards) - len(me.Hand.Cards)
	if diff <= 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: diff}.Apply(ctx)
}
