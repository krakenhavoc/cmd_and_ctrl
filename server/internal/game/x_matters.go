package game

// x_matters.go — the catalog's one declaration about the announced X
// (#810).
//
// CR 601.2b / 602.2b make X a number the caster announces, and CR
// 107.3 leaves 0 a legal announcement wherever the printed text does
// not forbid it. The engine is right to accept X=0 and does: this
// file changes nothing about what a cast or an activation is allowed
// to be.
//
// What it answers is a different question, asked only by the bot's
// legal-move enumerator (ADR 0033 §1): does the card DO anything at
// X=0? "{X}: Look at the top X cards" does not, and a free move that
// does nothing comes back as a move the moment it resolves — a
// repeatable sequence of optional actions no player is ever made to
// keep repeating (CR 732.2a). The catalog knows the answer and
// nothing else can derive it, so the catalog says it once, on the
// Spec, and `internal/legal` reads it through here.
//
// See effects.Spec.XMatters for what to declare and
// effects/x_matters_guard_test.go for the lint that makes sure it is
// declared.

// CatalogXMatters is the catalog hook the effects package wires at
// init, mirroring CatalogAdditionalLandPlays and its neighbours. Nil
// (no catalog) means no card declares it.
var CatalogXMatters func(oracleID string) bool

// XMattersFor reports whether everything the card does scales with
// the announced X — so an announcement of X=0 does nothing at all.
func XMattersFor(oracleID string) bool {
	if CatalogXMatters == nil || oracleID == "" {
		return false
	}
	return CatalogXMatters(oracleID)
}
