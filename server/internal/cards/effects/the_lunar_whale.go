package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Lunar Whale — Legendary Artifact — Vehicle {3}{U}, 3/5:
//
//	"Flying
//	 You may look at the top card of your library any time.
//	 As long as The Lunar Whale attacked this turn, you may play the top
//	 card of your library.
//	 Crew 1"
//
// Looking at the top card is LibraryTopVisible (The Reality Chip); the
// permission to play it is a gated standing permission, live only while
// this object has attacked this turn (the per-object attack history, so a
// Whale that left and came back has not). Crew is the Vehicle's ordinary
// crew ability.
func init() {
	Register(Spec{
		OracleID:          "13b4aeac-4530-4e09-a547-548b4884bb2f",
		Name:              "The Lunar Whale",
		Completeness:      CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		GatedCastPermissions: []game.CastPermissionGate{{
			Permission: PlayFromTopOfYourLibrary(game.PermissionFilter{},
				"Play a land or cast a spell from the top of your library (The Lunar Whale)"),
			Condition: func(g *game.Game, _, source uuid.UUID) bool {
				return g.AttackedThisTurn(source)
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("The Lunar Whale"),
		}},
	})
}
