package effects

// Disallow — Instant {1}{U}{U} (EDHREC rank 1824):
//
//	"Counter target spell, activated ability, or triggered ability.
//	 (Mana abilities can't be targeted.)"
//
// Cancel with a Stifle stapled on, and the three-way clause is the
// whole card: `TargetSpellOrAbility` with no narrowing enumerates the
// spell cards in the stack zone and the ability items in `StackMeta`
// into one legal set, and `CounterTarget` has discriminated on
// `StackItem.Kind` since S13.1.
//
// The three halves of the printed sentence are three shapes, not one:
// a countered SPELL goes to its owner's graveyard (CR 701.5a, through
// the shared stack-exit primitive, so a flashed-back one is exiled and
// a countered commander gets its CR 903.9 offer); a countered
// ACTIVATED or TRIGGERED ability goes NOWHERE (CR 701.6a — the item is
// deleted, its source permanent stays in play, and nothing it paid is
// refunded); and a MANA ability cannot be reached at all, because
// CR 605.3b keeps it off the stack, so the parenthetical needs no code.
//
// The caveat this card shipped with ("only spells can be countered —
// an activated or triggered ability on the stack can't be picked") is
// gone as of #1211: it named a missing target clause, and the clause
// exists now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88b51e15-6630-4e14-a6b8-db0aa12e34ef",
		Name:         "Disallow",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrAbility("target spell, activated ability, or triggered ability"),
		OnResolve:    counterTheTargetSpell,
	})
}
