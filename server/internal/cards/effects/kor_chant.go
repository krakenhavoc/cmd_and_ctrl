package effects

// Kor Chant — Instant {2}{W}:
//
//	"All damage that would be dealt this turn to target creature you
//	 control by a source of your choice is dealt to another target
//	 creature instead."
//
// ADR 0108 §9 (#1905): korRedirection. The targets are chosen as it is
// cast and the source as it resolves (the ruling); the damage dealt to
// the second target is still the original source's.
//
// No simplifications.
func init() {
	Register(korRedirection("5b3f6817-5d7a-4d83-ad1d-df75b4e1970b", "Kor Chant"))
}
