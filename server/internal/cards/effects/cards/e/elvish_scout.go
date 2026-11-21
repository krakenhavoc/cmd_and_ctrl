package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Scout — Creature — Elf Scout {G}, 1/1:
//
//	"{G}, {T}: Untap target attacking creature you control. Prevent all
//	 combat damage that would be dealt to and dealt by it this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Ebony Horse's ability on a
// creature, so the {T} needs it to have been under your control since
// your turn began (CR 302.6).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d39a492e-a599-4283-bfc1-8e99065a9303",
		Name:         "Elvish Scout",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.UntapAttackerToAndByRow(
			"{G}, {T}: Untap target attacking creature you control. Prevent all combat damage that would be dealt to and dealt by it this turn.",
			effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			effects.TargetCreature("target attacking creature you control", effects.AttackingCreature(), effects.YouControl()))},
	})
}
