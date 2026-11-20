package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pia, Determined Rebuilder — Legendary Creature — Human Artificer
// {2}{R}, 2/2:
//
//	"When Pia enters, create a 1/1 colorless Thopter artifact creature
//	 token with flying.
//	 {5}{R}: Target creature gets +X/+0 until end of turn, where X is
//	 the number of artifacts you control."
//
// X is counted as the ability resolves, so an artifact destroyed in
// response shrinks the pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a49c0c64-ea87-44f4-9acf-e7181962aeb2",
		Name:         "Pia, Determined Rebuilder",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pia, Determined Rebuilder — create a 1/1 Thopter with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{R}: Target creature gets +X/+0 until end of turn, where X is the number of artifacts you control",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Cost:    effects.ManaCost("{5}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.CountControlled(g, item.Controller, game.Card.IsArtifact)
				for _, t := range ctx.LegalTargets() {
					return effects.BoostUntilEOT{Target: t.ID, Power: x, Label: "Pia, Determined Rebuilder — +X/+0"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
