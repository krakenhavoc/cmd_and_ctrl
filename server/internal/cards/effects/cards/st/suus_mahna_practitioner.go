package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suus Mahna Practitioner — Creature — Vulcan Officer {2}{U}, 3/2
// (Star Trek):
//
//	Strange new worlds — {4}{U}, {T}: Tap target creature. This ability
//	costs {X} less to activate, where X is the number of differently
//	named lands you control.
//
// "Strange new worlds" is an ability word with no rules meaning. The
// discount is the ability's own cost clause (ActivatedAbility.CostModifiers,
// Mobilized District's shape), counted as the cost is determined (CR 602.2b,
// 601.2f): each land the activator controls counted once per name. It takes
// off generic mana only, so the {U} is always paid. The {T} keeps summoning
// sickness in force (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7d22ef6-a926-4964-bfd6-f7345b20e74e",
		Name:         "Suus Mahna Practitioner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Strange new worlds — {4}{U}, {T}: Tap target creature. This ability costs {X} less to activate, where X is the number of differently named lands you control.",
			Cost:    effects.Plus(effects.ManaCost("{4}{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.TrekB9DifferentlyNamedPermanentsYouControl(effects.Land()),
					"This ability costs {X} less to activate, where X is the number of differently named lands you control."),
			},
			Effect: effects.TapTheFirstLegalTarget,
		}},
	})
}
