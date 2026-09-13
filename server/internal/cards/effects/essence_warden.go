package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Warden — Creature — Elf Shaman {G}, 1/1 (EDHREC rank 1180):
//
//	"Whenever another creature enters, you gain 1 life."
//
// Soul Warden in green, word for word. ANY creature, under anyone's
// control; "another" excludes its own entry only; one trigger per
// creature, so a five-token batch is five triggers, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ca2a89e-7032-4864-b4e9-66f3178f90ab",
		Name:         "Essence Warden",
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
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Essence Warden — you gain 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
