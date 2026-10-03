package effects

// Kor Dirge — Instant {2}{B}:
//
//	"All damage that would be dealt this turn to target creature you
//	 control by a source of your choice is dealt to another target
//	 creature instead."
//
// ADR 0108 §9 (#1905): Kor Chant's text in black (korRedirection).
//
// No simplifications.
func init() {
	Register(korRedirection("44ae1c25-8622-4a58-92e0-12d76f084718", "Kor Dirge"))
}
