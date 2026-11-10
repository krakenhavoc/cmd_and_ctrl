package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra Avenger — Creature — Angel {W}{W}, 3/3:
//
//	"You can't cast Serra Avenger during your first, second, or third
//	 turns of the game.
//	 Flying, vigilance"
//
// Read off the caster's own count of turns begun (TurnsBegunFor), so an
// extra turn counts and a cast on another player's turn is untouched,
// as the ruling says.
func init() {
	Register(Spec{
		OracleID:        "b91b65b3-b35c-460e-a49b-333000a9a519",
		Name:            "Serra Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		CastCondition: func(g *game.Game, controller uuid.UUID, _ game.Card) bool {
			return !cmdrB1IsYourFirstThreeTurns(g, controller)
		},
		CastConditionLabel: "You can't cast Serra Avenger during your first, second, or third turns of the game.",
	})
}
