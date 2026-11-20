package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Songstitcher — Creature — Human Cleric {W}:
//
//	"{1}{W}: Prevent all combat damage that would be dealt this turn by target attacking creature with flying."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted attacker, pinned as the ability resolves (CR 400.7). Flying is
// a targeting requirement only: an attacker that loses flying after the
// ability resolves is still the shield's source.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eec72bcf-ccbc-43fc-8bdd-7bf4faa1fad7",
		Name:         "Songstitcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldAgainstTargetsRow(
			"{1}{W}: Prevent all combat damage that would be dealt this turn by target attacking creature with flying.",
			effects.ManaCost("{1}{W}"),
			effects.TargetCreature("target attacking creature with flying", effects.AttackingCreature(), effects.HasKeyword("flying")), true)},
	})
}
