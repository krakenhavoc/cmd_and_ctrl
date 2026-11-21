package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secluded Starforge — Land:
//
//	"{T}: Add {C}.
//	 {2}, {T}, Tap X untapped artifacts you control: Target creature
//	 gets +X/+0 until end of turn. Activate only as a sorcery.
//	 {5}, {T}: Create a 2/2 colorless Robot artifact creature token."
//
// #1421 closes the middle ability's one missing shape: TapXUntapped
// marks the tap clause CountFromX, so the number of artifacts picked
// is the announcement carried on StackItem.XValue. The effect reads
// that same immutable X through Context.X; it never recounts the
// artifacts after they have tapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "69f55a7c-6ddf-412e-b63b-b395731a1ff2",
		Name:          "Secluded Starforge",
		Completeness:  effects.CompletenessFull,
		XMatters:      true,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{2}, {T}, Tap X untapped artifacts you control: Target creature gets +X/+0 until end of turn.",
				Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0)),
				Cost:         effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.TapXUntapped("X untapped artifacts you control", effects.Artifact())),
				SorcerySpeed: true,
				Targets:      effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: ctx.X(), Label: "Secluded Starforge"}.Apply(ctx)
				},
			},
			{
				Label:   "{5}, {T}: Create a 2/2 colorless Robot artifact creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("2/2 colorless Robot artifact"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
