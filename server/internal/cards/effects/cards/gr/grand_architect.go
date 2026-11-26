package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grand Architect — Creature — Vedalken Artificer {1}{U}{U}, 1/3 (Scars of
// Mirrodin, #5577):
//
//	"Other blue creatures you control get +1/+1.
//	 {U}: Target artifact creature becomes blue until end of turn.
//	 Tap an untapped blue creature you control: Add {C}{C}. Spend this mana
//	 only to cast artifact spells or activate abilities of artifacts."
//
// The anthem is Mindwrack Liege's (layer 7c over the other blue creatures
// you control). The colour change is a layer-5 effect until end of turn
// (CR 613.1e). The mana ability's cost is tapping an untapped blue creature
// you control, the Architect itself included, which is not the {T} symbol
// (CR 302.6) and so ignores summoning sickness; the mana carries Vedalken
// Engineer's cast-or-activate artifact restriction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7af1d7f9-058a-4ccb-96b7-fc391b615c09",
		Name:         "Grand Architect",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(effects.G2OtherCreaturesYouControlOfColor("U"), 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Target artifact creature becomes blue until end of turn.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Gap2B05TargetBecomesColor(effects.NewContext(g, item), []string{"U"}, "Grand Architect — blue until end of turn")
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				TapOthers: &game.TapOthersCost{
					Count:  1,
					Filter: effects.TargetPermanent("an untapped blue creature you control", effects.Creature(), effects.OfColor("U")),
					Label:  "an untapped blue creature you control",
				},
			},
			Produced: "{C}{C}",
			Label:    "Tap an untapped blue creature you control: Add {C}{C}. Spend this mana only to cast artifact spells or activate abilities of artifacts.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
	})
}
