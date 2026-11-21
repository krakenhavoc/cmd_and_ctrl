package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crashing Drawbridge — Artifact Creature — Wall {2}, 0/4:
//
//	"Defender
//	 {T}: Creatures you control gain haste until end of turn."
//
// Defender rides PrintedKeywords like any other combat keyword. The
// tap ability is an ordinary S32 turn-scoped keyword grant
// (`GrantKeywordUntilEOT`) over "creatures you control" — the same
// `And(Creature(), YouControl())` match Overrun uses — snapshotted
// once at activation per CR 611.2c, so a creature cast afterward this
// turn doesn't get haste from it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "328d5ef0-b25e-4f2a-80fe-35c6ae419e8a",
		Name:            "Crashing Drawbridge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Creatures you control gain haste until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Keywords: []string{"haste"}}, Answers: game.AnswerCombatGrant},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"haste"},
					Label:    "Crashing Drawbridge — haste",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
