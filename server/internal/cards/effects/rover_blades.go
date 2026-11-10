package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rover Blades — Artifact — Equipment Vehicle {3}, 2/2 (Aetherdrift,
// #4464):
//
//	"Double strike
//	 Equipped creature has double strike.
//	 Equip {4}
//	 Crew 2 (Tap any number of creatures you control with total power 2
//	 or more: This Vehicle becomes an artifact creature until end of turn.
//	 Creatures can't be attached to other permanents.)"
//
// Fireshrieker's grant and equip, with Crew from the engine (CR 702.122a).
// The last reminder sentence is the rule the engine already keeps for an
// Equipment that is also a creature (CR 301.5c, game/attach.go): a crewed
// Rover Blades can't equip, as an Equipment creature without reconfigure
// may not, and one that is attached when it becomes a creature falls off.
// An uncrewed one (it is no creature) equips as any Equipment does, and
// stays attached when the crew effect ends. Its own double strike is the
// printed keyword, in effect while it is a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0394846e-bc27-4d67-988e-b08a389690de",
		Name:            "Rover Blades",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{
			GrantToAttached("double strike"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Rover Blades"),
			},
		},
	})
}
