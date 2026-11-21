package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoard Hauler — Artifact — Vehicle {3}{R}, 5/5 (Streets of New Capenna):
//
//	"Trample
//	 Whenever this Vehicle deals combat damage to a player, create a
//	 Treasure token for each artifact they control.
//	 Crew 3"
//
// "They" is the player the Vehicle dealt the damage to, read off the
// damage event; the artifacts are counted as the trigger resolves, so a
// Treasure sacrificed in response is not counted. The Treasures are the
// Vehicle's controller's. A player with no artifacts gives none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2578e8e4-a93c-4904-9037-92a1529c2c67",
		Name:            "Hoard Hauler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Hoard Hauler — create a Treasure token for each artifact they control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(effects.PermanentsControlledByMatching(g, item.Trigger.Event.Target, effects.Artifact()))
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Hoard Hauler"),
		}},
	})
}
