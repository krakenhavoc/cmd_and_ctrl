package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diplomatic Escort — Creature — Human Spellshaper {1}{U}, 1/1 (Mercadian
// Masques, S82 OLD batch 700, #5619):
//
//	"{U}, {T}, Discard a card: Counter target spell or ability that targets a
//	 creature."
//
// Siren Stormtamer's counter (TargetSpellOrAbility) behind the Spellshapers'
// cost: {U}, {T} and a discard, all paid as it is activated, the {T} being a
// creature's so the Escort waits out summoning sickness. The pick is a spell
// or an ability with a creature on the battlefield among its targets
// (S82OldB700ItemTargetsACreature), and a target that has gone by resolution
// fizzles the counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55cc14d0-11ed-444d-884a-cfb692ba632b",
		Name:         "Diplomatic Escort",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Discard a card: Counter target spell or ability that targets a creature.",
			Cost:    effects.NemSpellshaperCost("{U}"),
			Targets: effects.TargetSpellOrAbility("target spell or ability that targets a creature", effects.S82OldB700ItemTargetsACreature),
			Effect:  effects.S82OldB700CounterTheTargetItem,
		}},
	})
}
