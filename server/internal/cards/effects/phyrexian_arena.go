package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Arena — Enchantment for {1}{B}{B}:
//
//	"At the beginning of your upkeep, you lose 1 life and draw a
//	card."
//
// S19 sub-PR 5: the first "your upkeep" trigger on the auto-fire
// pipeline. EventBeginUpkeep carries the active player in Actor;
// AppliesTo gates on Actor == Controller so the Arena only fires on
// its own controller's upkeep (not every player's). Mandatory, so
// Build runs inline — lose 1 life, then draw 1.
func init() {
	Register(Spec{
		OracleID: "ee579a32-a048-4335-b966-231ba731cdea",
		Name:     "Phyrexian Arena",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				ctx := NewContext(g, nil)
				_ = g.ChangePlayerLifeForEffect(source.InstanceID, source.Controller, -1)
				_ = DrawCards{Player: source.Controller, N: 1}.Apply(ctx)
				return nil
			},
		}},
	})
}
