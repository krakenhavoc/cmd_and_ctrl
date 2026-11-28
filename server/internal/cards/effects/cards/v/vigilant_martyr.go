package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigilant Martyr — Creature — Human Cleric {W}, 1/1 (Mirage, S82 OLD batch 812,
// #5633):
//
//	"Sacrifice this creature: Regenerate target creature.
//	 {W}{W}, {T}, Sacrifice this creature: Counter target spell that targets an
//	 enchantment."
//
// Two abilities with a sacrifice cost each. The regeneration shield goes on the
// target as the ability resolves (CR 701.19a), whether or not the Martyr is
// still around; the counter needs a spell on the stack with an enchantment on
// the battlefield among its targets, and does nothing if that spell is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "122dfe9f-e0e2-4f69-9637-289c956fd54a",
		Name:         "Vigilant Martyr",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice this creature: Regenerate target creature.",
				Cost:    effects.SacrificeThis(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.RegenerateTheTargetPermanent,
			},
			{
				Label:   "{W}{W}, {T}, Sacrifice this creature: Counter target spell that targets an enchantment.",
				Cost:    effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost(), effects.SacrificeThis()),
				Targets: effects.TargetSpell("target spell that targets an enchantment", effects.S82OldB812SpellTargetsAnEnchantment()),
				Effect:  effects.S82OldB700CounterTheTargetItem,
			},
		},
	})
}
