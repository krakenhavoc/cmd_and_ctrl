package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloomfang Mauler — Creature — Nightmare {5}{B}{B}, 5/5 (March of the Machine):
//
//	"Swampcycling {2}
//	 Backup 2 (When this creature enters, put two +1/+1 counters on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Menace"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// menace until end of turn; the card has it itself, declared in PrintedKeywords. Swampcycling is Typecycling (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5437f560-7f61-4326-be8e-15d74bf4e6ec",
		Name:            "Gloomfang Mauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Gloomfang Mauler", 2, "", "menace"),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Swampcycling", "{2}", "a Swamp card", effects.IsLandWithSubtype("Swamp")),
		},
	})
}
