package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Core Prowler — Artifact Creature — Phyrexian Horror {4}, 2/2 (Mirrodin Besieged,
// #5576):
//
//	"Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)
//	 When this creature dies, proliferate. (Choose any number of
//	 permanents and/or players, then give each another counter of each
//	 kind already there.)"
//
// Infect is the engine's damage tail (CR 702.90b); the dies trigger is
// Blightbelly Rat's proliferate, chosen as it resolves (CR 701.34a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ad44727-41c3-471c-b038-1bf43c5f296a",
		Name:            "Core Prowler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Core Prowler — proliferate", effects.Do(effects.Proliferate{})),
		},
	})
}
