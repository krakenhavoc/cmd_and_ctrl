package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Butcher of Malakir — Creature — Vampire Warrior {5}{B}{B}, 5/4:
//
//	"Flying"
//	"Whenever this creature or another creature you control dies, each
//	 opponent sacrifices a creature of their choice."
//
// Grave Pact on a body, and the "THIS creature or another" wording
// matters: the Butcher's own death triggers it, so removal aimed at
// the Butcher still costs each opponent a creature. Compare Grave
// Pact, which says only "a creature you control" — the enchantment
// can't die to creature removal in the first place, so it never needed
// the clause.
//
// Expensive enough that it's a payoff rather than an engine, but it
// closes games in a way the enchantment can't: it attacks.
func init() {
	Register(Spec{
		OracleID:        "a85197ab-dc94-4b72-9716-8dbdbbe90ff8",
		Name:            "Butcher of Malakir",
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				// "This creature OR another creature you control" —
				// the source's own death counts, and diedCreature has
				// already resolved the card post-move so its
				// controller is still readable.
				return ok && dead.Controller == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Butcher of Malakir — each opponent sacrifices a creature",
					func(g *game.Game, item *game.StackItem) error {
						return EachPlayerSacrifices{
							ExceptController: true,
							Match:            Creature(),
							Label:            "a creature",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
