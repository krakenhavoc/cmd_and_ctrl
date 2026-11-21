package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omen of the Hunt — Enchantment {2}{G} (Theros Beyond Death, #5514):
//
//	"Flash
//	 When this enchantment enters, you may search your library for a
//	 basic land card, put it onto the battlefield tapped, then
//	 shuffle.
//	 {2}{G}, Sacrifice this enchantment: Scry 2."
//
// Omen of the Sea's frame: flash, an enters trigger, and a sacrifice-
// for-scry-2 activated ability, each Omen paying {2} and one mana of its
// own color. The enters trigger is Heaped Harvest's search: "you may" is
// the search prompt's decline, and the land enters tapped. Nothing is
// revealed: the card doesn't say to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "53295b03-717d-481d-9d5c-212c5f30bbbf",
		Name:            "Omen of the Hunt",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Omen of the Hunt — you may search for a basic land, tapped", func(g *game.Game, item *game.StackItem) error {
				return effects.B07SearchBasicOntoBattlefield(g, item, item.Controller, true, true, "Omen of the Hunt — a basic land card, onto the battlefield tapped")
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this enchantment: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
