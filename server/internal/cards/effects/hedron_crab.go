package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedron Crab — Creature — Crab {U}, 0/2:
//
//	"Landfall — Whenever a land you control enters, target player
//	 mills three cards."
//
// One mana for three cards a turn, forever, and the reason a mill
// deck plays lands at all. Landfall is the ordinary ETB-watch-plus-
// IsLand shape Lotus Cobra established.
//
// TARGET PLAYER, not "each opponent" and not "that player" — a Crab
// can mill its own controller, which is not a misprint but the point
// in a self-mill deck. The trigger therefore carries a real target
// clause, re-checked on resolution (CR 608.2b) like any other, so a
// player who left the table between the land drop and the resolution
// fizzles the trigger rather than erroring.
//
// The 0/2 body is real and relevant: it blocks, which is what buys
// the Crab the turns it needs.
func init() {
	Register(Spec{
		OracleID: "7216f974-3c84-40ef-b904-82019900c204",
		Name:     "Hedron Crab",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Targets: TargetPlayer("target player"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Hedron Crab — target player mills three (landfall)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
							return nil
						}
						if !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						return MillCards{Player: item.Targets[0].ID, N: 3}.Apply(ctx)
					})
			},
		}},
	})
}
