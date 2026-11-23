package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Castle Garenbrig — Land:
//
//	"This land enters tapped unless you control a Forest.
//	 {T}: Add {G}.
//	 {2}{G}{G}, {T}: Add six {G}. Spend this mana only to cast creature
//	 spells or activate abilities of creatures."
//
// Castle Vantress's entry clause and mana ability with the six-{G}
// ability on top. The restriction is one ManaRestrictAnyOf tag: cast a
// creature spell, or pay for an activated ability whose source is a
// creature (CR 106.6). The activation half reads the source's
// effective types.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de75e5dd-8a52-406c-b55c-96d686885500",
		Name:         "Castle Garenbrig",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B06EntersTappedUnlessLandType("forest")},
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{2}{G}{G}"},
				Produced: "{G}{G}{G}{G}{G}{G}",
				Label:    "{2}{G}{G}, {T}: Add six {G}. Spend this mana only to cast creature spells or activate abilities of creatures",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
				)},
			},
		},
	})
}
