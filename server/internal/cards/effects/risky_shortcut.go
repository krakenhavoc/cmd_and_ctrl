package effects

// Risky Shortcut — Sorcery {2}{B} (EDHREC rank 3647):
//
//	"Draw two cards. Each player loses 2 life."
//
// Night's Whisper at a table: the same two cards, and everyone pays
// the two life rather than only the caster. The draws land first,
// then every live player — the caster included — loses 2, life LOSS
// rather than damage, so no prevention shield and no lifelink sees
// it, and a player at 2 is dead at the next state check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7c55495-6aca-4d17-b61a-cab8fa38b9f4",
		Name:         "Risky Shortcut",
		Completeness: CompletenessFull,
		OnResolve:    b35DrawTwoThenEachPlayerLosesTwo,
	})
}
