package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hexbane Tortoise — Creature — Turtle {2}{G}, 3/2 (Dominaria United,
// #5487, seam #5658):
//
//	"Ward {2} (Whenever this creature becomes the target of a spell or
//	 ability an opponent controls, counter it unless that player pays {2}.)
//	 Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)"
//
// Ward {2} is the shared CR 702.21a trigger.
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3b4097f-b84f-4645-81f1-018366e40eeb",
		Name:            "Hexbane Tortoise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"ward", game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Hexbane Tortoise — ward {2}"),
		},
	})
}
