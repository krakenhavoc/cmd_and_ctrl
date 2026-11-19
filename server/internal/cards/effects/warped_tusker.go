package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warped Tusker — Creature — Eldrazi Boar Beast {7}, 6/8 (Modern Horizons 3):
//
//	"Reach
//	 When you cast or cycle Warped Tusker, create a 0/1 colorless Eldrazi Spawn
//	 creature token with "Sacrifice this token: Add {C}."
//	 Cycling {2}{G}"
//
// "Cast or cycle" is one printed ability, registered as the two rows the engine
// needs: the cast half watches the stack, the cycle half the graveyard.
func init() {
	const label = "Warped Tusker — create a 0/1 colorless Eldrazi Spawn creature token"
	Register(Spec{
		OracleID:        "0a994401-d37c-4e30-9410-422e7ae5a715",
		Name:            "Warped Tusker",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell(label, gapB51EldraziSpawn),
			InGraveyard(On(game.EventCycle, Self, label, gapB51EldraziSpawn)),
		},
	})
}
