package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "0a994401-d37c-4e30-9410-422e7ae5a715",
		Name:            "Warped Tusker",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(label, effects.GapB51EldraziSpawn),
			effects.InGraveyard(effects.On(game.EventCycle, effects.Self, label, effects.GapB51EldraziSpawn)),
		},
	})
}
