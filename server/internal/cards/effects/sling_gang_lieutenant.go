package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sling-Gang Lieutenant — Creature — Goblin {3}{B}, 1/1 (Modern Horizons 3):
//
//	"When this creature enters, create two 1/1 red Goblin creature tokens.
//	 Sacrifice a Goblin: Target player loses 1 life and you gain 1 life."
func init() {
	Register(Spec{
		OracleID:     "4e93b23c-a7f7-4bdd-bbca-0dc48bb5c223",
		Name:         "Sling-Gang Lieutenant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sling-Gang Lieutenant — create two 1/1 red Goblin creature tokens",
				Do(CreateToken{Template: RedGoblinToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Target player loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost:    game.AbilityCost{SacrificeOther: b12SacrificeAGoblin()},
			Targets: TargetPlayer("target player"),
			Effect:  drainTargetOpponentOne,
		}},
	})
}
