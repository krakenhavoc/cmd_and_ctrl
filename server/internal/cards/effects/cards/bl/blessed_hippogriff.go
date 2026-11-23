package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Hippogriff // Tyr's Blessing — an Adventure creature (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	Blessed Hippogriff {3}{W}, Creature — Hippogriff, 2/3
//	  "Flying
//	   Whenever this creature attacks, target attacking creature without flying
//	   gains flying until end of turn."
//	Tyr's Blessing {W}, Instant — Adventure
//	  "Target creature gains indestructible until end of turn."
//
// The attack trigger is Bast's shape (one trigger per declaration, the target
// an attacking creature picked as it goes on the stack, here one that has no
// flying at that moment). Indestructible is a canonical keyword, so the grant
// is real.
//
// No simplification.
func init() {
	const oracleID = "3372b994-47de-4e89-9bfc-76e1cb3f1cfa"
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Blessed Hippogriff",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Blessed Hippogriff — target attacking creature without flying gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
							return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Blessed Hippogriff — flying"}.Apply(ctx)
						}
						return nil
					}),
				effects.TargetCreature("target attacking creature without flying", effects.AttackingCreature(), effects.WithoutKeyword("flying"))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Tyr's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"indestructible"}, Label: "Tyr's Blessing — indestructible"}.Apply(ctx)
			}
			return nil
		},
	})
}
