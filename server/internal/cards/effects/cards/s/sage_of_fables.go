package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage of Fables — Creature — Merfolk Wizard {2}{U}, 2/2 (Lost Caverns of
// Ixalan Commander, #5368):
//
//	"Each other Wizard creature you control enters with an additional
//	 +1/+1 counter on it.
//	 {2}, Remove a +1/+1 counter from a creature you control: Draw a card."
//
// The first line is a CR 614 entry replacement (Grumgully's shape), so
// Hardened Scales and Doubling Season see the counter. The cost removes
// the counter from a creature chosen as the cost is paid (Ray Fillet).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f53da056-97be-41ad-8cfc-0d92e17dcd7b",
		Name:         "Sage of Fables",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.Cmdr2B23OtherWizardsEnterWithACounter("Sage of Fables: enters with an additional +1/+1 counter"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Remove a +1/+1 counter from a creature you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.RemoveCountersFrom(game.CounterPlusOne, 1, "a creature you control", effects.Creature())),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
