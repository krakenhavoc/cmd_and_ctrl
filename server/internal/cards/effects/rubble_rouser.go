package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rubble Rouser — Creature — Dwarf Sorcerer {2}{R}, 1/4 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, you may discard a card. If you do, draw a card.
//	 {T}, Exile a card from your graveyard: Add {R}. When you do, this creature
//	 deals 1 damage to each opponent."
//
// The loot is the shared discard-then-draw prompt (Iceman and Firestar's
// loot): the draw only follows a discard that happened. The second ability is
// a mana ability (CR 605.1a) whose exile is a cost the activator picks
// (ExileCardsFromGraveyard, the exile-from-graveyard cost component on a mana
// ability) and whose damage is the rider after "Add {R}"
// (PingOpponents, Zhur-Taa Druid's declaration): it runs in the same atomic
// resolution, so the damage is not a separate stack object the way the
// printed reflexive "When you do" is (CR 603.12), and nobody can respond
// between the mana and the damage. That is the catalog's reading of a mana
// ability's damage clause.
//
// No simplification beyond that timing.
func init() {
	const lootLabel = "Rubble Rouser — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "f35aa497-bd0f-4642-8713-820b9f014da5",
		Name:         "Rubble Rouser",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, lootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:          ManaAbilityCost{Tap: true, ExileCards: ExileCardsFromGraveyard(1, "a card", nil)},
			Produced:      "{R}",
			Label:         "{T}, Exile a card from your graveyard: Add {R}. When you do, this creature deals 1 damage to each opponent",
			PingOpponents: 1,
		}},
	})
}
