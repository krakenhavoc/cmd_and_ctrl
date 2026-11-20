package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shabraz, the Skyshark — Legendary Creature — Shark Bird {3}{W}{U}, 3/3
// (Commander 2020, #5513):
//
//	"Partner with Brallin, Skyshark Rider (When this creature enters, target
//	 player may put Brallin into their hand from their library, then
//	 shuffle.)
//	 Flying
//	 Whenever you draw a card, put a +1/+1 counter on Shabraz and you gain 1
//	 life.
//	 {W/U}: Target Human gains flying until end of turn."
//
// Partner with is the CR 702.124j entry search (partner_with.go). The draw
// trigger is Clinquant Skymage's, once per card drawn, with the life gain
// beside the counter. The hybrid activation is Air Marshal's targeted
// flying grant narrowed to a Human.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7599fe9a-b6a5-491b-a308-aef3bf1e9daf",
		Name:            "Shabraz, the Skyshark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			PartnerWith("Shabraz, the Skyshark", "Brallin, Skyshark Rider"),
			WheneverYouDraw("Shabraz, the Skyshark — put a +1/+1 counter on it and you gain 1 life",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W/U}: Target Human gains flying until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    ManaCost("{W/U}"),
			Targets: TargetCreature("target Human", Subtype("Human")),
			Effect:  ebTargetCreatureGainsUntilEOT("flying", "Shabraz, the Skyshark — flying until end of turn"),
		}},
	})
}
