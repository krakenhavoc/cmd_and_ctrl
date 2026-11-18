package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sky Crier — Creature — Bird Citizen {1}{W}, 1/1 (Streets of New Capenna,
// #5491):
//
//	"Flying, lifelink
//	 {3}{W}: You and target opponent each draw a card."
//
// Flying and lifelink ride PrintedKeywords. The draw is the shared "you and
// target opponent each draw" body (Flumph, Teyo): the controller draws first,
// then the still-legal target; the ability's controller is the activator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d88c2375-65ec-4909-aef3-ee1a658a07cc",
		Name:            "Sky Crier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: You and target opponent each draw a card.",
			Cost:    ManaCost("{3}{W}"),
			Targets: TargetPlayer("target opponent", Opponent()),
			Purpose: game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})},
			Effect:  youAndTargetOpponentEachDraw(),
		}},
	})
}
