package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arenson's Aura — Enchantment {2}{W} (Ice Age, S82 OLD, #5636):
//
//	"{W}, Sacrifice an enchantment: Destroy target enchantment.
//	 {3}{U}{U}: Counter target enchantment spell."
//
// Teferi's Care's two abilities: the sacrifice is a cost (the Aura itself qualifies), the counter
// targets a spell on the stack and is re-checked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "465843dc-57d0-46fd-ac47-238723034563",
		Name:         "Arenson's Aura",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, Sacrifice an enchantment: Destroy target enchantment.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeN(1, "an enchantment", effects.Enchantment())),
				Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
				Effect:  effects.DestroyChosenPermanent,
			},
			{
				Label:   "{3}{U}{U}: Counter target enchantment spell.",
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Targets: effects.TargetSpell("target enchantment spell", effects.Enchantment()),
				Effect:  effects.S82OldB700CounterTheTargetItem,
			},
		},
	})
}
