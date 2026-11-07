package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gastal Thrillroller — Artifact — Vehicle {2}{R}, 4/2 (Aetherdrift,
// #4464):
//
//	"Trample, haste
//	 When this Vehicle enters, it becomes an artifact creature until end
//	 of turn.
//	 Crew 2
//	 {2}{R}, Discard a card: Return this card from your graveyard to the
//	 battlefield with a finality counter on it. Activate only as a
//	 sorcery."
//
// The enters trigger is crew's effect (BecomeCreatureUntilEOT) on the
// Vehicle itself, pinned to the object that entered, so a Thrillroller
// that left and came back in response is not animated (CR 400.7). With
// haste it can attack the turn it arrives (CR 702.10b). Crew is the
// engine's (CR 702.122a).
//
// The graveyard ability is Uchbenbak's return (Grim Repriser's zone)
// with a discard in its cost: the card comes back with the finality
// counter as it enters, and a card that left the graveyard in response
// stays where it is. It enters as a fresh object, so its own enters
// trigger animates it again. Sorcery timing is CR 602.5d; a permanent
// with a finality counter that would go to a graveyard is exiled instead
// (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86d67952-19b8-4bb8-b7d7-b2751e86591c",
		Name:            "Gastal Thrillroller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gastal Thrillroller — it becomes an artifact creature until end of turn",
				Do(BecomeCreatureUntilEOT{Label: "Gastal Thrillroller — becomes an artifact creature"})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Gastal Thrillroller"),
			},
			{
				Label:        "{2}{R}, Discard a card: Return this card from your graveyard to the battlefield with a finality counter on it. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{2}{R}"), DiscardACard()),
				Zones:        []game.ZoneKind{game.ZoneGraveyard},
				SorcerySpeed: true,
				Effect:       returnThisFromGraveyardWithAFinalityCounter,
			},
		},
	})
}
