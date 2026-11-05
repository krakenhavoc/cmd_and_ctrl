package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocksteady, Mutant Marauder — Legendary Creature — Rhino Mutant {2}{G},
// 3/3 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Partner with Bebop, Skull & Crossbones (When this creature enters,
//	 target player may put Bebop into their hand from their library,
//	 then shuffle.)
//	 Trample
//	 Whenever another nontoken creature you control enters, put a +1/+1
//	 counter on target creature."
//
// Partner with is the shared CR 702.124j pair: the commander pairing is
// read off the oracle text by internal/deck, and the entry search is
// PartnerWith. The counter trigger is Soul of the Harvest's condition
// (another nontoken creature entering under your control, read through
// the layered view) with Hunting Moa's counter on any target creature,
// Rocksteady and the creature that entered included. A target gone by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "506a972d-d226-44dc-95b7-a6e7a63d76cd",
		Name:            "Rocksteady, Mutant Marauder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Rocksteady, Mutant Marauder", "Bebop, Skull & Crossbones"),
			Targeting(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b15AnotherNontokenCreatureYouControlEntered(ev, source, g)
			}, "Rocksteady, Mutant Marauder — put a +1/+1 counter on target creature", putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature")),
		},
	})
}
