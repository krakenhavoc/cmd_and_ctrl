package effects

// Everlasting Torment — Enchantment {2}{B/R}:
//
//	"Players can't gain life.
//	 Damage can't be prevented.
//	 All damage is dealt as though its source had wither. (A source with
//	 wither deals damage to creatures in the form of -1/-1 counters.)"
//
// Three battlefield statics, each read live while the enchantment is on
// the battlefield and has its abilities (CR 613.1f):
//
//   - CR 119.7's "can't gain life" for every player (ADR 0107 §5);
//   - CR 615.12's "damage can't be prevented" for every damage event (ADR
//     0107 §5). Its ruling: "Spells and abilities that replace or redirect
//     damage aren't affected", and they aren't: only a prevention effect is
//     settled without running;
//   - ADR 0108 §10's "dealt as though its source had wither" (CR 120.3d,
//     702.80a, 609.4): damage to a creature is -1/-1 counters, put on it
//     by the source's controller, from any source in any zone (its ruling:
//     "Wither works everywhere"). The source gains no ability, and damage
//     to a player or a planeswalker is unchanged.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "de41081e-713c-4005-90f7-65d9426208f1",
		Name:                  "Everlasting Torment",
		Completeness:          CompletenessFull,
		CantGainLife:          PlayersCantGainLife(),
		DamageCantBePrevented: DamageCantBePreventedStatic(),
		DamageAsThough:        AllDamageAsThoughWither(),
	})
}
