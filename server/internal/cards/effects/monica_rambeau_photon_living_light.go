package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Monica Rambeau // Photon, Living Light — a double-faced Legendary
// Creature (Marvel Super Heroes, #4455):
//
//	Monica Rambeau {R}{W}, Legendary Creature — Human Hero, 1/3
//	  "Flying, prowess
//	   {2}{R}{W}{W}: Transform Monica Rambeau. Activate only as a
//	   sorcery."
//
//	Photon, Living Light, Legendary Creature — Elemental Hero, 3/5
//	  "Flying, hexproof, prowess
//	   Whenever you cast a noncreature spell, put a +1/+1 counter on
//	   each other creature you control."
//
// Prowess is the engine keyword (game/prowess.go). The counters are
// Mikaeus, the Lunarch's: one placement per creature, the set snapshotted
// when the trigger resolves (CR 608.2), so a creature that left in
// response gets none and Photon herself gets none.
//
// No simplification.
func init() {
	const oracle = "8dd929cb-0fe1-4cb3-bc8d-96ca37f194f3"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Monica Rambeau",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		Activated: []ActivatedAbility{{
			Label:        "{2}{R}{W}{W}: Transform Monica Rambeau. Activate only as a sorcery.",
			Cost:         ManaCost("{2}{R}{W}{W}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Photon, Living Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "hexproof", "prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Photon — put a +1/+1 counter on each other creature you control",
				putCounterOnEachOtherCreatureYouControl),
		},
	})
}
