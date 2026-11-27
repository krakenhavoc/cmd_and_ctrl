package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivy Dancer — Creature — Dryad Shaman {2}{G}, 1/2 (Ravnica: City of Guilds,
// S82 OLD batch 574, #5600):
//
//	"{T}: Target creature gains forestwalk until end of turn. (It can't be
//	 blocked as long as defending player controls a Forest.)"
//
// Zombie Trailblazer's grant: forestwalk through the layered view
// until end of turn, on the creature targeted as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2e081be-87ef-457a-a5cb-2c2ff10cbcf5",
		Name:         "Ivy Dancer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains forestwalk until end of turn. (It can't be blocked as long as defending player controls a Forest.)",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "forestwalk")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.GrantKeywordUntilEOT{Target: effects.FirstLegalBattlefieldTarget(ctx), Keywords: []string{"forestwalk"},
					Label: "Ivy Dancer — forestwalk"}.Apply(ctx)
			},
		}},
	})
}
