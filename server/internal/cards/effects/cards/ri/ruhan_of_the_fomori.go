package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruhan of the Fomori — Legendary Creature — Giant Warrior {1}{U}{R}{W}, 7/7
// (Secret Lair Drop):
//
//	"At the beginning of combat on your turn, choose an opponent at random.
//	 Ruhan attacks that player this combat if able."
//
// The random pick is the game's seeded draw over the living opponents; the
// requirement is Territorial Hellkite's attack-that-player-if-able effect,
// lasting until end of combat (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82d109ea-3421-46eb-9bbf-f7f28b548ea6",
		Name:         "Ruhan of the Fomori",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Ruhan of the Fomori — choose an opponent at random; Ruhan attacks that player this combat if able",
				effects.S82OldB021RuhanChoose),
		},
	})
}
