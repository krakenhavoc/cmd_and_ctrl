package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sakura-Tribe Elder — Creature — Snake Shaman {1}{G}, 1/1:
//
//	"Sacrifice Sakura-Tribe Elder: Search your library for a basic
//	land card, put it onto the battlefield tapped, then shuffle."
//
// The classic "chump block, then ramp anyway" body — the ability has
// no mana or tap component, so it can be activated at any time
// including after blocking, which is the whole trick.
//
// No tap cost means summoning sickness never applies: the Elder can
// sacrifice itself the turn it lands.
func init() {
	Register(Spec{
		OracleID: "e3afc704-220f-498f-9eaa-0821b17dc24c",
		Name:     "Sakura-Tribe Elder",
		Activated: []ActivatedAbility{{
			Label: "Sacrifice Sakura-Tribe Elder: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:  SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     IsBasicLand,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Sakura-Tribe Elder — a basic land",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
