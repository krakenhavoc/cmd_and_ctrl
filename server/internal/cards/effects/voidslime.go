package effects

// Voidslime — Instant {G}{U}{U} (EDHREC rank 4544):
//
//	"Counter target spell, activated ability, or triggered ability.
//	 (Mana abilities can't be targeted.)"
//
// Disallow in green-blue, printed nine years earlier and word for word
// the same clause. Both are `TargetSpellOrAbility` with no narrowing
// at all — the widest thing the stack can be pointed at — and both
// resolve through the one primitive that has discriminated on
// `StackItem.Kind` since S13.1.
//
// The two cards are deliberately written the same way rather than one
// delegating to the other: they differ only in colour and cost, both
// of which come off the printed card, and a clone family whose members
// are one call each is cheaper to read than an indirection.
//
// No simplification. (See Disallow for the CR 701.5c note on where a
// countered ability goes, which is nowhere.)
func init() {
	Register(Spec{
		OracleID:     "c4851768-e210-4aaf-935e-715942e198f9",
		Name:         "Voidslime",
		Completeness: CompletenessFull,
		Targets:      TargetSpellOrAbility("target spell, activated ability, or triggered ability"),
		OnResolve:    counterTheTargetSpell,
	})
}
