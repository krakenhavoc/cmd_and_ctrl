package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tangle Angler — Creature — Phyrexian Horror {3}{G}, 1/5 (Scars of Mirrodin,
// S82 OLD batch 471, #5577):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)
//	 {G}: Target creature blocks this creature this turn if able."
//
// Infect is printed. The block requirement is Vortex Elemental's
// (S82OldB378TargetBlocksThis): it names this creature as the object it is now
// and asks nothing of a blocker that can't legally block it (CR 509.1c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "887a9f2f-17b4-4a20-a464-9665cde6f293",
		Name:            "Tangle Angler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Target creature blocks this creature this turn if able.",
			Cost:    effects.ManaCost("{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB378TargetBlocksThis,
		}},
	})
}
