package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slobad, Goblin Tinkerer — Legendary Creature — Goblin Artificer {1}{R},
// 1/2:
//
//	"Sacrifice an artifact: Target artifact gains indestructible until end
//	 of turn."
//
// The artifact is sacrificed as a cost, so it can be Slobad's own target
// only if it is a different artifact (the sacrificed one is gone before
// targets are checked at resolution; a target that left does nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f2d6cf4-b8f4-440f-ac08-ca47ada868ac",
		Name:         "Slobad, Goblin Tinkerer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: Target artifact gains indestructible until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")), game.AnswerSurvive),
			Cost:    effects.SacrificeN(1, "an artifact", effects.Artifact()),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect: effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"},
					Label: "Slobad, Goblin Tinkerer — indestructible until end of turn"}.Apply(ctx)
			})),
		}},
	})
}
