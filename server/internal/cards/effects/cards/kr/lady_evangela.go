package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lady Evangela — Legendary Creature — Human Cleric {W}{U}{B}:
//
//	"{W}{B}, {T}: Prevent all combat damage that would be dealt by target creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted creature, pinned as the ability resolves (CR 400.7). The {T}
// is the tap symbol, so summoning sickness applies (CR 302.6).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8800d672-424b-4a7b-886f-7eb9d7a56cfe",
		Name:         "Lady Evangela",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.ShieldAgainstTargetsRow(
			"{W}{B}, {T}: Prevent all combat damage that would be dealt by target creature this turn.",
			effects.Plus(effects.ManaCost("{W}{B}"), effects.TapCost()), effects.TargetCreature("target creature"), true)},
	})
}
