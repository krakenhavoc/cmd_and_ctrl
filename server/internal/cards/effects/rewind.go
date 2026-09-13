package effects

// Rewind — Instant {2}{U}{U} (EDHREC rank 1013):
//
//	"Counter target spell. Untap up to four lands."
//
// The free Counterspell: four mana in, four lands back. Counterspell's
// clause, then the refund, in printed order.
//
// Sandbox simplification, declared — the Snap posture: "untap up to
// four lands" prints no "target" and no "you control", so on paper
// it is a resolution-time choice among every land at the table. The
// engine has no pick-a-permanent prompt for a spell, so this untaps
// the first four TAPPED lands the caster controls, in battlefield
// order. That is the outcome a player chooses in every real game, it
// is one the printed card allows, and it is never stronger — only
// less controllable.
func init() {
	Register(Spec{
		OracleID:     "bb27bfdf-fe8d-45bd-ad62-8118dce06eda",
		Name:         "Rewind",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You can't choose which lands untap — it automatically untaps the first four tapped lands you control and can never untap an opponent's lands."},
		Targets:      TargetSpell("target spell"),
		OnResolve:    b09CounterThenUntapLands(4),
	})
}
