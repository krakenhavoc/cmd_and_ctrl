package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izoni, Thousand-Eyed — Legendary Creature — Elf Shaman {2}{B}{B}{G}{G},
// 2/3:
//
//	"Undergrowth — When Izoni enters, create a 1/1 black and green Insect
//	 creature token for each creature card in your graveyard.
//	 {B}{G}, Sacrifice another creature: You gain 1 life and draw a card."
//
// Undergrowth is an ability word, so the first line is a plain enters
// trigger: the creature cards in its controller's graveyard are counted as
// it resolves (a card in a graveyard has its printed types, CR 108.2).
// The second is a sacrifice outlet whose cost names another creature, then
// the life gain and the draw, in that order.
func init() {
	const name = "Izoni, Thousand-Eyed"
	Register(Spec{
		OracleID:     "5a210b32-6e1f-4578-bd8b-4e481b761159",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(name+" — create a 1/1 black and green Insect creature token for each creature card in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					n := creatureCardsInGraveyard(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 black and green Insect"), N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}, Sacrifice another creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1, Draws: 1},
			Cost:    Plus(ManaCost("{B}{G}"), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  Do(GainLife{Amount: 1}, DrawCards{N: 1}),
		}},
	})
}
