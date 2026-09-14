package effects

// Flow of Knowledge — Instant {4}{U} (EDHREC rank 3540):
//
//	"Draw a card for each Island you control, then discard two cards."
//
// The mono-blue refill. The Islands are counted as the spell
// resolves, by effective subtype, so a land that has been made an
// Island counts; the discard is a two-card prompt of the caster's
// choice that opens after the draws, so the drawn cards are among
// the options. With no Islands the caster draws nothing and still
// discards two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b65f20de-52fa-4904-bbfe-7ba53ccd8ae5",
		Name:         "Flow of Knowledge",
		Completeness: CompletenessFull,
		OnResolve:    b33DrawPerIslandThenDiscardTwo,
	})
}
