package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// DECLARED GAP, weaker than printed: an Aura fetched this way enters
// unattached and stays that way, because the attach target is chosen by
// the casting spell and a fetch has no spell (Zur the Enchanter's gap).
func init() {
	Register(Spec{
		OracleID:     "3b319540-4db1-43d3-8f4d-f5a276ad78e4",
		Name:         "Guardian Sunmare",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"An Aura fetched this way enters unattached and stays that way — you don't get to choose what it enchants."},
		Activated:    []ActivatedAbility{Saddle(4)},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Guardian Sunmare — ward {2}"),
			AttacksWhileSaddled("Guardian Sunmare — search for a nonland permanent with mana value 3 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return !c.IsLand() && !c.IsInstant() && !c.IsSorcery() && c.ManaValue() <= 3
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Reason:  "Guardian Sunmare — a nonland permanent card with mana value 3 or less, onto the battlefield",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
