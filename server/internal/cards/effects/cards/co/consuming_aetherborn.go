package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consuming Aetherborn — Creature — Aetherborn Vampire {3}{B}, 2/2 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Lifelink"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// lifelink until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d472308-5bc7-4a3c-81b5-0a757f00d9e2",
		Name:            "Consuming Aetherborn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Consuming Aetherborn", 1, "", "lifelink"),
		},
	})
}
