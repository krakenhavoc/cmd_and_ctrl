package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canopy Tactician — Creature — Elf Warrior {3}{G}, 3/3 (EDHREC rank
// 3214):
//
//	"Other Elves you control get +1/+1.
//	 {T}: Add {G}{G}{G}."
//
// An Elf lord that is also a three-mana dork. The anthem is the
// shared tribal builder over OTHER Elves the controller controls;
// the mana ability is a plain tap for three green, summoning-sick
// like any creature's tap ability (CR 302.6 — the engine enforces
// it, the spec does not declare it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b20d6f4-6322-4435-92d5-acaae74774f4",
		Name:         "Canopy Tactician",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Elf"}, Others: true, YoursOnly: true}, 1, 1),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
	})
}
