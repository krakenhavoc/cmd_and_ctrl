package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Organic Avulsion Unit — Artifact Creature — Borg {2}{B}, 2/3 (Star Trek,
// #4973):
//
//	"{2}: Artifact creatures you control gain menace until end of turn."
//
// The grant pins its affected set as the ability resolves (CR 611.2c): an
// artifact creature that arrives, or a permanent that becomes an artifact
// creature, after it resolves does not gain menace. Types are read off the
// layered view, so an animated artifact counts.
//
// No simplification.
func init() {
	const name = "Organic Avulsion Unit"
	effects.Register(effects.Spec{
		OracleID:     "f1e28022-b06b-4efb-9f85-c36d8302ef99",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Artifact creatures you control gain menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.YouControl(), effects.Artifact(), effects.Creature()),
					Keywords: []string{"menace"},
					Label:    name + " — menace until end of turn",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
