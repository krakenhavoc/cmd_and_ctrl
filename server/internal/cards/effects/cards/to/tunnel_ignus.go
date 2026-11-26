package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tunnel Ignus — Creature — Elemental {1}{R}, 2/1 (Scars of Mirrodin, S82 OLD
// batch 471, #5577):
//
//	"Whenever a land enters under an opponent's control, if that player had
//	 another land enter the battlefield under their control this turn, this
//	 creature deals 3 damage to that player."
//
// Zo-Zu the Punisher's land watcher, narrowed to an opponent's land and gated
// on the turn's entry tally (PlayerTurnTally.LandsEntered): two or more means
// another land entered before this one. The player is read as the trigger is
// built, so a land that has since left is still attributed to its controller.
// The "if" can't become false (the tally only grows), so resolution needs no
// second look.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41353001-eba4-45b5-8dd2-09d49b63b6b1",
		Name:         "Tunnel Ignus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				p, ok := effects.B43ALandAnOpponentControlsEntered(ev, source, g)
				return ok && g.TurnTallyFor(p).LandsEntered >= 2
			},
			Key: "Tunnel Ignus — 3 damage to that player",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Tunnel Ignus — 3 damage to that player")
				item.Params.Player, _ = effects.B43ALandAnOpponentControlsEntered(ev, source, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDamage{Source: item.SourceCardID, Target: item.Params.Player, Amount: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
