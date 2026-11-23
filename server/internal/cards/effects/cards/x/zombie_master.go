package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zombie Master — Creature — Zombie {1}{B}{B}, 2/3:
//
//	"Other Zombie creatures have swampwalk.
//	 Other Zombies have "{B}: Regenerate this permanent.""
//
// Both clauses reach every other Zombie on the table, whoever controls it. The
// swampwalk is a layer 6 keyword grant (TribalKeywordGrant); the regeneration
// is a granted activated ability (ADR 0093), so "this permanent" is the Zombie
// that has it and only its controller can pay for it.
//
// No simplification.
const zombieMasterRegenerateGrant = "zombie-master/regenerate"

func init() {
	filter := effects.TribeFilter{Tribes: []string{"Zombie"}, Others: true}
	effects.Register(effects.Spec{
		OracleID:     "5446c92f-ff22-4e9b-a2f6-e64c8560c1e0",
		Name:         "Zombie Master",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: zombieMasterRegenerateGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{B}: Regenerate this permanent.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{B}: Regenerate this permanent.",
		}},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(filter, "swampwalk"),
			effects.TribalAbilityGrant(filter, zombieMasterRegenerateGrant),
		},
	})
}
