package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethersworn Shieldmage — Artifact Creature — Vedalken Wizard {1}{W}{U},
// 2/2:
//
//	"Flash
//	 When this creature enters, prevent all damage that would be dealt
//	 to artifact creatures this turn."
//
// #2045's recipient set with two queries that must BOTH match: an
// artifact that is also a creature, whoever controls it (the Shieldmage
// itself included), read as the damage would be dealt (CR 611.2c). A
// Vehicle that becomes crewed later in the turn is protected; an
// artifact creature that stops being a creature is not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "ee988017-fc7e-4d8a-8f5c-0a7e57a8d050",
		Name:            "Ethersworn Shieldmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ethersworn Shieldmage — prevent all damage that would be dealt to artifact creatures this turn",
				Do(PreventDamageFromSource{Protect: ShieldPermanents(QueryTypes("artifact"), QueryTypes("creature"))})),
		},
	})
}
