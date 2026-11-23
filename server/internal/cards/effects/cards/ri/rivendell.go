package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rivendell — Legendary Land (EDHREC rank 951):
//
//	"Rivendell enters tapped unless you control a legendary creature.
//	 {T}: Add {U}.
//	 {1}{U}, {T}: Scry 2. Activate only if you control a legendary
//	 creature."
//
// The blue sibling of Mines of Moria and The Shire: the same "unless
// you control a legendary creature" entry, read post-layer, and a
// plain {U}. The scry's "Activate only if you control a legendary
// creature" is its activation condition (CR 602.1b, #743), the same
// test the entry makes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2550099d-b3e2-4eb6-9f36-0fc412828ca6",
		Name:         "Rivendell",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTappedUnless(effects.B08ControlsLegendaryCreature)},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{U}, {T}: Scry 2. Activate only if you control a legendary creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue, Scry: 2},
			Cost:      effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(1, effects.MatchLegendaryCreature),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Scry{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
