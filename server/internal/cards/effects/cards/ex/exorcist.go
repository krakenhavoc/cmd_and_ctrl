package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exorcist — Creature — Human Cleric {W}{W}, 1/1 (The Dark, #5638):
//
//	"{1}{W}, {T}: Destroy target black creature."
//
// Royal Assassin's row with a mana component and a colour clause in
// place of "tapped". OfColor reads the creature's current colours, at
// announce and again at resolution, so a creature that stops being
// black in response is no longer a legal target and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a1ca3e2-e1d1-4525-a07e-13d5ba1ac431",
		Name:         "Exorcist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Destroy target black creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target black creature", effects.OfColor("B")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
