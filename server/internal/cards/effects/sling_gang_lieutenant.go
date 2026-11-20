package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sling-Gang Lieutenant — Creature — Goblin {3}{B}, 1/1 (Modern Horizons, #5519):
//
//	"When this creature enters, create two 1/1 red Goblin creature tokens.
//	 Sacrifice a Goblin: Target player loses 1 life and you gain 1 life."
//
// Beetleback Chief's two Goblins and Pashalik Mons' "Sacrifice a Goblin"
// cost (any Goblin you control, the Lieutenant included) on Gnawing
// Zombie's drain.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4e93b23c-a7f7-4bdd-bbca-0dc48bb5c223",
		Name:         "Sling-Gang Lieutenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sling-Gang Lieutenant — create two 1/1 red Goblins", Do(CreateToken{Template: RedGoblinToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Target player loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost:    b25SacrificeAGoblin(),
			Targets: TargetPlayer("target player"),
			Effect:  drainTargetOpponentOne,
		}},
	})
}
