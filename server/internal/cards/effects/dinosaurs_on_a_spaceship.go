package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dinosaurs on a Spaceship — Creature — Dinosaur {5}{R}{W}, 6/6:
//
//	"Vigilance, trample
//	 Other Dinosaurs you control get +1/+1 and have vigilance and trample.
//	 Suspend 4—{3}{R}{W}
//	 Whenever a time counter is removed from this card while it's exiled,
//	 create a 2/2 red and white Dinosaur creature token with flying and
//	 haste."
//
// #2466: the trigger works from exile (ZoneExile, as suspend's own two
// do) and fires once per time counter, so the upkeep countdown's four
// ticks make four Dinosaurs, and a Clockspinning that takes three off at
// once makes three. The last tick's token and the free cast are two
// separate triggers. The lord half is the shared tribal builders.
//
// No simplifications.
func init() {
	dinos := TribeFilter{Tribes: []string{"Dinosaur"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:        "e3b4314b-e7a2-449a-af5a-817c8260adf1",
		Name:            "Dinosaurs on a Spaceship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Static: []game.StaticAbility{
			TribalAnthem(dinos, 1, 1),
			TribalKeywordGrant(dinos, "vigilance"),
			TribalKeywordGrant(dinos, "trample"),
		},
		SpecialActions: []game.SpecialAction{Suspend(4, "{3}{R}{W}")},
		Triggered: []game.TriggeredAbility{
			WheneverACounterIsRemovedFromThis(game.CounterTime,
				"Dinosaurs on a Spaceship — create a 2/2 red and white Dinosaur with flying and haste",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("2/2 red and white Dinosaur with flying and haste"), N: 1}.Apply(NewContext(g, item))
				}, game.ZoneExile),
		},
	})
}
