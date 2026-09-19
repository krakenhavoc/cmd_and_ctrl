package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Ruinous Blast — {4}{W} Legendary Sorcery:
//
//	"(You may cast a legendary sorcery only if you control a legendary
//	 creature or planeswalker.)
//	 Exile all nonland permanents that aren't legendary."
//
// CR 307.6, and the other half of #760: a restriction printed on the
// SPELL rather than on somebody's permanent. It is the same gate
// (ADR 0073 §7) with the other of its two sources — which is why the
// two issues are one ADR: a design that only did statics would have
// had to invent a second mechanism a week later for this.
//
// Checked once, at announce, and NEVER at resolution. A legendary
// creature that dies with the Blast on the stack does not counter it:
// CR 307.6 restricts the CAST, and nothing in the rules re-checks it.
// That is why this is a CastCondition rather than something the
// OnResolve asks.
//
// The reminder text is the label, verbatim, because the label is what
// the refusal shows the player. LegendarySorceryLabel keeps the ten
// legendary sorceries from spelling it ten ways.
func init() {
	Register(Spec{
		OracleID:           "978e0d87-3ff2-4a73-916c-ff0dc0ab2797",
		Name:               "Urza's Ruinous Blast",
		Completeness:       CompletenessFull,
		CastCondition:      LegendarySorcery(),
		CastConditionLabel: LegendarySorceryLabel,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ExileAllMatching{
				Match: And(Not(Land()), Not(Legendary())),
			}.Apply(ctx)
		},
	})
}
