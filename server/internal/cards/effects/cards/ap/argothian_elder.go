package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argothian Elder — Creature — Elf Druid {3}{G}, 2/2 (Urza's Saga):
//
//	"{T}: Untap two target lands."
//
// Ley Weaver's ability without the partner: exactly two target lands,
// any player's. A land that has left by resolution is skipped
// (CR 608.2b), and the {T} cost waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b57f06e4-da97-4473-9fbf-f66e9129502a",
		Name:         "Argothian Elder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Untap two target lands.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("two target lands", effects.Land()).WithCount(2, 2),
			Effect:  effects.UntapEachLegalTarget,
		}},
	})
}
