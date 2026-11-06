package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Purple Pentapus — Creature — Octopus Starfish {B}, 1/1 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"When this creature enters, surveil 1.
//	 {2}{B}, Tap an untapped creature you control: Return this card from
//	 your graveyard to the battlefield tapped."
//
// The return is Reassembling Skeleton's (returnThisFromGraveyardTapped),
// activated only from the graveyard. Tapping a creature is not the {T}
// symbol (CR 302.6), so one that arrived this turn may pay it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37d2e83d-f5fd-48d2-b091-3c9167522530",
		Name:         "Purple Pentapus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Purple Pentapus — surveil 1", Do(Surveil{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Tap an untapped creature you control: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), TapAnUntapped("an untapped creature you control", Creature())),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisFromGraveyardTapped,
		}},
	})
}
