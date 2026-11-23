package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra Disciple — Creature — Bird Cleric {1}{W}, 1/1 (Dominaria, #5530):
//
//	"Flying, first strike
//	 Whenever you cast a historic spell, this creature gets +1/+1 until end of turn. (Artifacts, legendaries, and Sagas are historic.)"
//
// Jhoira's historic cast trigger (CR 700.6): +1/+1 until end of turn.
//
// No simplification.
func init() {
	const label = "Serra Disciple — +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "0f052d9f-5a40-4eb9-9d8a-e01dec1db191",
		Name:            "Serra Disciple",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastAHistoricSpell, label, effects.ThisGetsUntilEndOfTurn(1, 1, label)),
		},
	})
}
