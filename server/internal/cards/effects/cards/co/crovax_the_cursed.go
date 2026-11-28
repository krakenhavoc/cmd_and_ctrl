package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crovax the Cursed — Legendary Creature — Vampire Noble {2}{B}{B}, 0/0
// (Stronghold, S82 OLD batch 748, #5627):
//
//	"Crovax enters with four +1/+1 counters on it.
//	 At the beginning of your upkeep, you may sacrifice a creature. If you do,
//	 put a +1/+1 counter on Crovax. If you don't, remove a +1/+1 counter from
//	 Crovax.
//	 {B}: Crovax gains flying until end of turn."
//
// The entry counters are a self-replacement (CR 614.1c). The upkeep offer is
// Zoyowa Lava-Tongue's pay-or-else shape with a branch for each answer: a
// "yes" sacrifices a creature of the controller's choice and puts the counter
// on (CR 118.3), and a "no" -- or having no creature to sacrifice at all, which
// is not asked -- removes one. Both only while Crovax is still the permanent
// the trigger came from (CR 400.7). The flying is a grant until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1b74ef7-ab6b-4d40-8087-3f3a448b7e0a",
		Name:         "Crovax the Cursed",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB031SelfEntersWithCounters(game.CounterPlusOne, 4)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Crovax the Cursed — you may sacrifice a creature; if you do, put a +1/+1 counter on Crovax, if you don't, remove one", effects.CrovaxUpkeepOffer),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Crovax gains flying until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Keywords: []string{"flying"}}},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Crovax the Cursed — flying until end of turn"),
		}},
	})
}
