package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soratami Cloud Chariot — Artifact {5}:
//
//	"{2}: Target creature you control gains flying until end of turn.
//	 {2}: Prevent all combat damage that would be dealt to and dealt by
//	 target creature you control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the second ability is one
// to-and-by record (Mod.AndDealtBy) pinned to the target.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a46373a-b48a-4118-92e0-d856ce409690",
		Name:         "Soratami Cloud Chariot",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Target creature you control gains flying until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: ts[0].ID, Keywords: []string{"flying"}, Label: "Soratami Cloud Chariot — flying"}.Apply(ctx)
				},
			},
			effects.SourceShieldRow("{2}: Prevent all combat damage that would be dealt to and dealt by target creature you control this turn.",
				effects.ManaCost("{2}"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.ToAndByShield(effects.ShieldTheTarget, true)),
		},
	})
}
