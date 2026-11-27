package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mobile Fort — Artifact Creature — Wall {4} 0/6 (Urza's Saga, #5624):
//
//	"Defender (This creature can't attack.)
//	 {3}: This creature gets +3/-1 until end of turn and can attack this turn
//	 as though it didn't have defender. Activate only once each turn."
//
// The limit is the announce tally (CR 602.1b), so an activation still on the
// stack counts.
//
// No simplification.
func init() {
	const label = "{3}: This creature gets +3/-1 until end of turn and can attack this turn as though it didn't have defender. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "0547e426-e23e-4c60-ba6b-cfd4b5ff8485",
		Name:            "Mobile Fort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(3, -1, "Mobile Fort — +3/-1 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.CanAttackThisTurnAsThoughNoDefender("Mobile Fort — can attack this turn as though it didn't have defender")(g, item)
			},
		}},
	})
}
