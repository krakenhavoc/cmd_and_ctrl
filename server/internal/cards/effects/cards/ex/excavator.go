package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Excavator — Artifact {2} (Tempest, S82 OLD b752, #5628):
//
//	"{T}, Sacrifice a basic land: Target creature gains landwalk of each of the
//	 land types of the sacrificed land until end of turn. (It can't be blocked as
//	 long as defending player controls a land of any of those types.)"
//
// The sacrifice is paid as the ability is activated (CR 602.2b), after the target
// is chosen. The land types are read from the sacrificed land as it last existed
// on the battlefield (CR 603.10a; its leaves-the-battlefield event), so a Forest
// gives forestwalk, a land that was every basic type gives all five, and a land
// turned into a Mountain by Blood Moon gives mountainwalk. Each is one keyword the
// creature gains until the cleanup step (CR 514.2); a target that left does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2d5ebcd7-07c2-412e-b0d9-54e3a88895fd",
		Name:         "Excavator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a basic land: Target creature gains landwalk of each of the land types of the sacrificed land until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a basic land", effects.IsBasicLandCard())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				var keywords []string
				for _, ref := range ctx.Paid().SacrificedObjects {
					next, ok := effects.CardTheCostMovedToAPublicZone(g, ref)
					if !ok {
						continue
					}
					keywords = append(keywords, effects.OldB752BasicLandWalkKeywords(g, next)...)
				}
				if len(keywords) == 0 {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: keywords, Label: "Excavator — landwalk of the sacrificed land's types"}.Apply(ctx)
			},
		}},
	})
}
