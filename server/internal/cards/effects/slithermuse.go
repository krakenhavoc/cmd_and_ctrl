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
// resolution (CR 702.74b), which is what makes the sequence work: the
// creature genuinely enters, the sacrifice goes on the stack where it
// can be responded to, and the leave-trigger then goes on the stack
// above the empty board and draws. Folding the sacrifice into
// resolution would produce the same hand size and the wrong stack.
//
// Sandbox simplification: "choose an opponent" is auto-picked as the
// opponent holding the most cards, which is always the choice that
// draws the most. A player-choice prompt would need a trigger-side
// player picker.
func init() {
	Register(Spec{
		OracleID: "4b6512aa-535e-4edf-8797-43df2b8463de",
		Name:     "Slithermuse",
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{3}{U}"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Slithermuse — draw the hand-size difference",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						me := ctx.PlayerByID(item.Controller)
						if me == nil {
							return nil
						}
						best := 0
						for _, oppID := range ctx.Opponents() {
							opp := ctx.PlayerByID(oppID)
							if opp == nil {
								continue
							}
							if diff := len(opp.Hand.Cards) - len(me.Hand.Cards); diff > best {
								best = diff
							}
						}
						return DrawCards{Player: item.Controller, N: best}.Apply(ctx)
					})
			},
		}},
	})
}
