package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul's Attendant — Creature — Human Cleric {W}, 1/1 (EDHREC rank
// 833):
//
//	"Whenever another creature enters, you may gain 1 life."
//
// Soul Warden's twin, with the one printed difference kept: this one
// is "you may", so every creature at the table pops a yes/no for the
// Attendant's controller. Any controller's creature counts, tokens
// included; the Attendant's own entry does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "045a9d3d-c20d-427f-a77b-0bffc6f55526",
		Name:         "Soul's Attendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Soul's Attendant — gain 1 life?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Soul's Attendant — gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
