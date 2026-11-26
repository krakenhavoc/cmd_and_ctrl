package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Visage of Bolas — Artifact {4} (Hour of Devastation, #5534):
//
//	"When this artifact enters, you may search your library and/or graveyard
//	 for a card named Nicol Bolas, the Deceiver, reveal it, and put it into
//	 your hand. If you search your library this way, shuffle.
//	 {T}: Add {U}, {B}, or {R}."
//
// Basri's Aegis's named fetch as an enters trigger: the controller says which
// zones to search (the library is shuffled only if it was searched), reveals
// the card and puts it into their hand. The mana ability is Xander's Lounge's
// choice of one of three colors, and is a mana ability (CR 605.1a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21bb3c03-c1c0-40ec-8df4-8d0054c3742b",
		Name:         "Visage of Bolas",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Visage of Bolas — you may search your library and/or graveyard for a card named Nicol Bolas, the Deceiver, reveal it, and put it into your hand; if you search your library this way, shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.S82OldB105MayFetchNamed("Visage of Bolas", "Nicol Bolas, the Deceiver")(effects.NewContext(g, item))
				}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|B|R}",
			Label:    "Add {U}, {B}, or {R}",
		}},
	})
}
