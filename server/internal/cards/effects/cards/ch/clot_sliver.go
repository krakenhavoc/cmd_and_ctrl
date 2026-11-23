package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const clotSliverGrant = "clot-sliver/regenerate"

// Clot Sliver — Creature — Sliver {1}{B}, 1/1 (Commander Masters):
//
//	"All Slivers have "{2}: Regenerate this permanent.""
//
// Mindlash Sliver's shape: a bundle of one activated ability granted to every
// Sliver on the battlefield, the Clot Sliver itself and every player's
// included (TribalAbilityGrant without YoursOnly). The shield is the
// permanent's own: each Sliver regenerates itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a885729-0618-4a25-b18d-b0587f14718d",
		Name:         "Clot Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: clotSliverGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}: Regenerate this permanent.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				},
			}},
			Text: "{2}: Regenerate this permanent.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, clotSliverGrant),
		},
	})
}
