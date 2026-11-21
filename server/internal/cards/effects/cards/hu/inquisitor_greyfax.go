package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inquisitor Greyfax — Legendary Creature — Human Inquisitor {1}{W}{U}{B}, 3/3:
//
//	"Vigilance
//	 Unquestionable Wisdom — Other creatures you control get +1/+0 and have
//	 vigilance.
//	 Hunt for Heresy — {1}, {T}: Tap target creature an opponent controls.
//	 Investigate."
func init() {
	others := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID != source.InstanceID && target.IsCreature() && target.Controller == source.Controller
	}
	effects.Register(effects.Spec{
		OracleID:        "9439d6eb-4c10-451d-85c4-3752611c671c",
		Name:            "Inquisitor Greyfax",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.B16Anthem(others, 1, 0),
			effects.B16GrantKeywords(others, "vigilance"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Tap target creature an opponent controls. Investigate.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if ts := ctx.LegalTargets(); len(ts) > 0 {
					if err := (effects.TapTarget{Target: ts[0].ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.Investigate{N: 1}.Apply(ctx)
			},
		}},
	})
}
