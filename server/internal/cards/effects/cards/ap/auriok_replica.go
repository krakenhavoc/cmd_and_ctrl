package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Replica — Artifact Creature — Cleric {3}:
//
//	"{W}, Sacrifice this creature: Prevent all damage a source of your choice would deal to you this turn."
//
// ADR 0108 §7 (#1904): a shield against a source chosen as the ability
// resolves (CR 609.7a), protecting its controller from every instance of
// that source's damage this turn, not only the next.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "783a62c0-3b06-4250-a1cb-5cf4042611aa",
		Name:         "Auriok Replica",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.SourceShieldRow(
			"{W}, Sacrifice this creature: Prevent all damage a source of your choice would deal to you this turn.",
			effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()), nil,
			effects.PreventDamageFromChosenSource(effects.ShieldYou))},
	})
}
