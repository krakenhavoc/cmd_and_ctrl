package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coral Reef — Enchantment {U}{U}, (Homelands, S82 OLD b822, #5635):
//
//	"This enchantment enters with four polyp counters on it.
//	Sacrifice an Island: Put two polyp counters on this enchantment.
//	{U}, Tap an untapped blue creature you control, Remove a polyp counter from
//	this enchantment: Put a +0/+1 counter on target creature."
//
// Entering counters (B10EntersWithCounters), a sacrifice cost, and a tap-a-creature
// cost that is not the {T} symbol (CR 302.6; TapAnUntapped, Earthcraft's shape),
// so a blue creature that arrived this turn may pay it. The +0/+1 counter is a
// power/toughness counter (CR 122.1a).
//
// No simplification.
func init() {
	const polyp = "polyp"
	effects.Register(effects.Spec{
		OracleID:     "37a40a53-2668-4521-a786-c5747eedeeb2",
		Name:         "Coral Reef",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(polyp, 4, "Coral Reef: enters with four polyp counters on it"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice an Island: Put two polyp counters on this enchantment.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.SacrificeN(1, "an Island", effects.HasSubtype("Island")),
				Effect:  effects.Do(effects.CounterOnThis{Kind: polyp, N: 2}),
			},
			{
				Label: "{U}, Tap an untapped blue creature you control, Remove a polyp counter from this enchantment: Put a +0/+1 counter on target creature.",
				Cost: effects.Plus(effects.ManaCost("{U}"),
					effects.TapAnUntapped("an untapped blue creature you control", effects.Creature(), effects.OfColor("U")),
					effects.RemoveCountersFromThis(polyp, 1)),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.OldB822TargetGetsACounter("+0/+1"),
			},
		},
	})
}
