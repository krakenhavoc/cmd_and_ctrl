package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doppelgang — Sorcery {X}{X}{X}{G}{U} (EDHREC rank 2951):
//
//	"For each of X target permanents, create X tokens that are copies
//	 of that permanent."
//
// The Simic X-squared copier. The cost engine charges 3X generic
// (ParseCost counts every {X}); the target clause is "X target
// permanents" — any permanent, anyone's — with CountFromX, the
// Waterbender's Restoration / Crackle with Power hook that replaces
// the clause's count with the announced X before anything validates
// against it, so an X of 0 buys no targets and makes nothing. For
// each still-legal target (CR 608.2b per slot) the caster gets X
// token copies through CreateTokenCopy, which carries the copied
// card's oracle ID onto the tokens so their triggers, statics, mana
// and activated abilities all come along; a target that left in
// response is skipped and the rest are still copied. The targets are
// snapshotted before any token lands, so a copy is never itself
// copied.
//
// One declared simplification, weaker than printed — CreateTokenCopy's
// standing gap (Cackling Counterpart's): a token copy of a card whose
// enters-the-battlefield effect is an on-enter hook rather than a
// trigger skips that effect.
func init() {
	Register(Spec{
		OracleID:     "d04c6375-25dd-4882-b8b3-a3b0d3081f32",
		Name:         "Doppelgang",
		XMatters:     true,
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Token copies skip the enters-the-battlefield effect on some cards."},
		Targets:      b28DoppelgangTargets(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b28CopyEachLegalTargetXTimes(ctx)
		},
	})
}

// b28DoppelgangTargets is "X target permanents", any permanent each.
func b28DoppelgangTargets() *game.TargetSpec {
	spec := TargetPermanent("X target permanents")
	spec.CountFromX = true
	return spec
}
