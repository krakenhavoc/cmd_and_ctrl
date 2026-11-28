package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walking Wall — Artifact Creature — Wall {4}, 0/6 (Ice Age, S82 OLD, #5636):
//
//	"Defender
//	 {3}: This creature gets +3/-1 until end of turn and can attack this turn as
//	 though it didn't have defender. Activate only once each turn."
//
// Hightide Hermit's defender waiver until end of turn beside a +3/-1, gated by
// the once-per-turn tally (CR 602.5b).
//
// No simplification.
func init() {
	const label = "{3}: This creature gets +3/-1 until end of turn and can attack this turn as though it didn't have defender. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "909201e5-5599-442a-9e4f-ff61f187e957",
		Name:            "Walking Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.ManaCost("{3}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ThisGetsUntilEndOfTurn(3, -1, "Walking Wall — +3/-1 until end of turn")(g, item); err != nil {
					return err
				}
				return effects.CanAttackThisTurnAsThoughNoDefender("Walking Wall — can attack as though it didn't have defender")(g, item)
			},
		}},
	})
}
