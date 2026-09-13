package effects

// Unwind — Instant {2}{U} (EDHREC rank 1062):
//
//	"Counter target noncreature spell. Untap up to three lands."
//
// Rewind's little sibling: Negate's clause with a three-land refund.
// The noncreature gate is the target predicate, checked at announce
// and again at resolution.
//
// Sandbox simplification, declared — the Snap posture: "untap up to
// three lands" is a resolution-time choice among every land at the
// table on paper. With no pick-a-permanent prompt for a spell, this
// untaps the first three TAPPED lands the caster controls, in
// battlefield order — never an opponent's, never stronger than
// printed, only less controllable.
func init() {
	Register(Spec{
		OracleID:     "e38b8ecb-e7ae-474d-b6a4-29cc1aa8ccd9",
		Name:         "Unwind",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can't choose which lands untap — it automatically untaps the first three tapped lands you control and can never untap an opponent's lands."},
		Targets:      TargetSpell("target noncreature spell", Noncreature()),
		OnResolve:    b09CounterThenUntapLands(3),
	})
}
