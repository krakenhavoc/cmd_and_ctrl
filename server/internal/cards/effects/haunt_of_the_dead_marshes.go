package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunt of the Dead Marshes — Creature — Nightmare Elf {B}, 1/1
// (The Hobbit):
//
//	"When this creature enters, scry 1.
//	 {2}{B}: Return this card from your graveyard to the battlefield
//	 tapped. Activate only if you control a legendary creature."
//
// The enter trigger is a plain scry. The recursion is Reassembling
// Skeleton's graveyard ability (returnThisFromGraveyardTapped, which does
// nothing if the card has left the graveyard by resolution) behind an
// activation condition, ControlsAtLeast(1, MatchLegendaryCreature): it
// is checked when the ability is activated and greys the ability out
// otherwise, and the ability is not re-checked on resolution, as printed.
// The card returns as a new object, so entering scries again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b683611-e46f-4105-b336-014d1556c8de",
		Name:         "Haunt of the Dead Marshes",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Haunt of the Dead Marshes — scry 1", Do(Scry{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:     "{2}{B}: Return this card from your graveyard to the battlefield tapped. Activate only if you control a legendary creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{2}{B}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: ControlsAtLeast(1, MatchLegendaryCreature),
			Effect:    returnThisFromGraveyardTapped,
		}},
	})
}
