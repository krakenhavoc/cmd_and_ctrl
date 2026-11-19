package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Supplicant — Creature — Human Cleric {B}, 1/1:
//
//	"{T}, Sacrifice three Clerics: Search your graveyard, hand, and/or
//	 library for a card named Scion of Darkness and put it onto the
//	 battlefield. If you search your library this way, shuffle."
//
// Gate to the Afterlife's by-name search of all three zones.
func init() {
	Register(Spec{
		OracleID:     "dc7929cb-50a9-4662-a106-c415fc03a1a9",
		Name:         "Dark Supplicant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice three Clerics: Search your graveyard, hand, and/or library for a card named Scion of Darkness and put it onto the battlefield. If you search your library this way, shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeN(3, "three Clerics", HasSubtype("Cleric"))),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     func(c game.Card) bool { return c.Name == "Scion of Darkness" },
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Shuffle:       true,
					AlsoGraveyard: true,
					AlsoHand:      true,
					Reason:        "Dark Supplicant — a card named Scion of Darkness from your graveyard, hand or library",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
