package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Sunmare — Creature — Horse Mount {3}{W}{W}:
//
//	"Ward {2}
//	 Whenever this creature attacks while saddled, search your library
//	 for a nonland permanent card with mana value 3 or less, put it onto
//	 the battlefield, then shuffle.
//	 Saddle 4"
//
// Ward is the catalog helper (a triggered ability, ADR 0038 §7). The
// search is mandatory, not a "may", and finds any nonland permanent card
// of mana value 3 or less: creature, artifact, enchantment, planeswalker
// or battle.
//
// An Aura fetched this way chooses what it enchants as it enters
// (CR 303.4f, #3158); with nothing to enchant it stays in the library
// (CR 303.4g).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b319540-4db1-43d3-8f4d-f5a276ad78e4",
		Name:         "Guardian Sunmare",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(4)},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Guardian Sunmare — ward {2}"),
			effects.AttacksWhileSaddled("Guardian Sunmare — search for a nonland permanent with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return !c.IsLand() && !c.IsInstant() && !c.IsSorcery() && c.ManaValue() <= 3
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Reason:  "Guardian Sunmare — a nonland permanent card with mana value 3 or less, onto the battlefield",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
