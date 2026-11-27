package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leech Bonder — Creature — Merfolk Soldier {2}{U}, 3/3 (Shadowmoor, S82 OLD
// b532, #5591):
//
//	"This creature enters with two -1/-1 counters on it.
//	 {U}, {Q}: Move a counter from target creature onto a second target
//	 creature. ({Q} is the untap symbol.)"
//
// {Q} is AbilityCost.UntapSelf (CR 107.6). The move is Nesting Grounds'
// (Rikku's Steal): one counter, of the kind the activator picks when the first
// creature has several, put on the second creature and taken off the first
// only if it landed (CR 122.5). The second target is a different creature from
// the first (CR 115.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3f03d0e6-c7c0-484c-bb65-b8cbf0bd802b",
		Name:         "Leech Bonder",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB532MinusOneCountersAtEntry("Leech Bonder", 2)},
		Activated: []effects.ActivatedAbility{{
			Label: "{U}, {Q}: Move a counter from target creature onto a second target creature.",
			Cost:  effects.OldB532UntapCost("{U}"),
			Targets: effects.Clauses(
				effects.TargetCreature("target creature"),
				effects.Distinct(effects.TargetCreature("a second target creature")),
			),
			Effect: effects.NestingGroundsMove,
		}},
	})
}
