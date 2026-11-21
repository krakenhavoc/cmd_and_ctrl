package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cutthroat Centurion — Artifact Creature — Phyrexian Warrior {2}{B}, 2/2
// (Phyrexia: All Will Be One):
//
//	"Sacrifice another artifact or creature: This creature gets +2/+2 until
//	 end of turn. Activate only once each turn."
//
// The cost names a permanent other than the Centurion that is an artifact
// or a creature; the once-a-turn limit is read off the activation tally
// (CR 602.5b), so a second activation is never offered.
//
// No simplification.
func init() {
	const label = "Sacrifice another artifact or creature: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "5707c278-dab9-43bc-b3a3-a6b7923511f4",
		Name:         "Cutthroat Centurion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost:      effects.SacrificeAnotherN(1, "another artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Cutthroat Centurion — +2/+2"}.Apply(ctx)
			},
		}},
	})
}
