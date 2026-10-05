package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scion of Vitu-Ghazi — Creature — Elemental {3}{W}{W}, 4/4:
//
//	"When this creature enters, if you cast it from your hand, create
//	 a 1/1 white Bird creature token with flying, then populate.
//	 (Create a token that's a copy of a creature token you control.)"
//
// An intervening-if (CR 603.4): "you cast it from your hand" is read
// off the event log when the trigger fires and again as it resolves
// (b30CastFromHand), exactly as Wakening Sun's Avatar does. A
// reanimated, blinked or commander-zone Scion makes nothing. The Bird
// is made first, so it is a candidate for the populate.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a34c4d5c-6ecb-4da2-8734-a2f3d35cfe8b",
		Name:         "Scion of Vitu-Ghazi",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return b06SelfETB(ev, source, lki, g) && b30CastFromHand(g, source.InstanceID)
			}, "Scion of Vitu-Ghazi — a Bird, then populate", func(g *game.Game, item *game.StackItem) error {
				if !b30CastFromHand(g, item.SourceCardID) {
					return nil
				}
				ctx := NewContext(g, item)
				if err := (CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 1}).Apply(ctx); err != nil {
					return err
				}
				return Populate{}.Apply(ctx)
			}),
		},
	})
}
