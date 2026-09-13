package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garruk's Packleader — Creature — Beast {4}{G}, 4/4 (EDHREC rank
// 2027):
//
//	"Whenever another creature you control with power 3 or greater
//	 enters, you may draw a card."
//
// The big-creature deck's card engine. The condition reads the
// entering creature's CURRENT power at the moment it enters, so a
// 2/2 under an anthem counts and a 0/0 Hydra that has already got
// its counters does too; "another" excludes the Packleader itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13279222-422d-4447-9451-2463b0c714c6",
		Name:         "Garruk's Packleader",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.CurrentPower() >= 3
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Garruk's Packleader — draw a card?"},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Garruk's Packleader — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
