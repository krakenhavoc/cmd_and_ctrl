package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "37d2e83d-f5fd-48d2-b091-3c9167522530",
		Name:         "Purple Pentapus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Purple Pentapus — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Tap an untapped creature you control: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapAnUntapped("an untapped creature you control", effects.Creature())),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardTapped,
		}},
	})
}
