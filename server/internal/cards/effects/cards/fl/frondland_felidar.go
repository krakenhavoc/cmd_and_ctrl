package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Frondland Felidar — Creature — Cat Beast {2}{G}{W}, 3/5 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Vigilance
//	 Creatures you control with vigilance have \"{1}, {T}: Tap target
//	 creature.\""
//
// Hollowhead Sliver's ability grant to the creatures its controller
// controls that have vigilance, read through the layers (Jubilant
// Skybonder's shape), the Felidar itself included. The granted ability is
// an ordinary {T} cost, so summoning sickness applies to each creature
// that uses it (CR 302.6).
//
// No simplification.
func init() {
	const grant = "frondland-felidar/tap"
	effects.Register(effects.Spec{
		OracleID:        "bb504fb1-1d19-4d0b-958f-e8f2433e0575",
		Name:            "Frondland Felidar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			}},
			Text: "{1}, {T}: Tap target creature.",
		}},
		Static: []game.StaticAbility{
			effects.GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && game.HasKeyword(target, "vigilance")
			}, grant),
		},
	})
}
