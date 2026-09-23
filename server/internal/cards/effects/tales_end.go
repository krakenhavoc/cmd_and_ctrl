package effects

// Tale's End — Instant {1}{U} (EDHREC rank 4813):
//
//	"Counter target activated ability, triggered ability, or legendary
//	 spell."
//
// The Commander answer to a commander: it stops an ability the way
// Stifle does and, uniquely in the family, a SPELL — but only a
// legendary one. So its clause is the one card in #1211 that narrows
// the two halves of the stack DIFFERENTLY, and it is what the
// StackItemPredicate vocabulary exists for:
//
//	AnyStackItem(AnAbilityItem(), ASpellItem(Legendary()))
//
// Read left to right that is the printed sentence: any ability at all,
// or a spell that is legendary. Written as two predicates rather than
// one because the halves genuinely differ — "legendary" is a fact
// about a CARD and an ability has none, its source permanent being a
// separate object that stays where it is.
//
// Legendary is read post-layer (the `Legendary()` predicate goes
// through the effective type line), so a spell that is legendary only
// because something made it so is a legal target and one that has had
// the supertype stripped is not — and the check runs again at
// resolution (CR 608.2b), so a spell that stops being legendary in
// response makes this fizzle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8ebb6fe0-3ed1-4cc9-bcb1-5317de199efc",
		Name:         "Tale's End",
		Completeness: CompletenessFull,
		Targets: TargetSpellOrAbility(
			"target activated ability, triggered ability, or legendary spell",
			AnyStackItem(AnAbilityItem(), ASpellItem(Legendary())),
		),
		OnResolve: counterTheTargetSpell,
	})
}
