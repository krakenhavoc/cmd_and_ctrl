package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "2578e8e4-a93c-4904-9037-92a1529c2c67",
		Name:            "Hoard Hauler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Hoard Hauler — create a Treasure token for each artifact they control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(permanentsControlledByMatching(g, item.Trigger.Event.Target, Artifact()))
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: TreasureToken(), N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Hoard Hauler"),
		}},
	})
}
