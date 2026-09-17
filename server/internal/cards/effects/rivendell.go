package effects

import (
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
	Register(Spec{
		OracleID:     "2550099d-b3e2-4eb6-9f36-0fc412828ca6",
		Name:         "Rivendell",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(b08ControlsLegendaryCreature)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{1}{U}, {T}: Scry 2. Activate only if you control a legendary creature.",
			Cost:      Plus(ManaCost("{1}{U}"), TapCost()),
			Condition: ControlsAtLeast(1, MatchLegendaryCreature),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Scry{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
