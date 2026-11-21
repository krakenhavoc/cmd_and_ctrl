package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Felhide Petrifier — Creature — Minotaur Warrior {2}{B}, 2/3 (Journey into Nyx, #5558):
//
//	"Deathtouch
//	 Other Minotaur creatures you control have deathtouch."
//
// A lord's keyword half: other Minotaur creatures you control have
// deathtouch, read through the layers so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ad604ed-c645-4f94-8c5f-8e0ce66a85ff",
		Name:            "Felhide Petrifier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Minotaur"}, Others: true, YoursOnly: true}, "deathtouch"),
		},
	})
}
