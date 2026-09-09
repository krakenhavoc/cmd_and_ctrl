package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pitiless Plunderer — Creature — Human Pirate {3}{B}, 1/4:
//
//	"Whenever another creature you control dies, create a Treasure
//	token."
//
// The card that turns an aristocrats board into a mana engine: every
// sacrifice refunds a mana, so a sacrifice outlet with a mana cost
// becomes free and the loop keeps going. Treasure already cracks for
// mana through S21 sub-PR 1's sacrifice-cost mana abilities, so this
// needs nothing new.
//
// "ANOTHER creature you control" — the Plunderer's own death does not
// trigger it, unlike Zulaport Cutthroat's "this creature or another".
func init() {
	Register(Spec{
		OracleID: "a784481f-eccb-4112-bb38-04a659319660",
		Name:     "Pitiless Plunderer",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == source.InstanceID {
					return false // "another"
				}
				dead, ok := diedCreature(ev, g)
				return ok && dead.Controller == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Pitiless Plunderer — create a Treasure",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{
							Controller: item.Controller,
							Template:   TreasureToken(),
							N:          1,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
