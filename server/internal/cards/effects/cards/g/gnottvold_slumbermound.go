package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnottvold Slumbermound — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {3}{R}{G}{G}, {T}, Sacrifice this land: Destroy target land. Create a
//	 4/4 green Troll Warrior creature token with trample."
//
// The sacrifice is a cost, so the land is gone before the ability
// resolves; the target is another land (the land can't target itself and
// has left). If the target land is illegal on resolution the whole ability
// does nothing, token included (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e413d18-2c55-49be-9e49-54d170c978a2",
		Name:         "Gnottvold Slumbermound",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}{G}{G}, {T}, Sacrifice this land: Destroy target land. Create a 4/4 green Troll Warrior creature token with trample.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{3}{R}{G}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
					return err
				}
				return effects.CreateToken{Template: effects.TokenCard("4/4 green Troll Warrior with trample"), N: 1}.Apply(ctx)
			},
		}},
	})
}
