package effects

// Mirror Gallery — Artifact {5}:
//
//	"The "legend rule" doesn't apply."
//
// The exemption covers every player's permanents (CR 704.5j), read live
// off the battlefield by the legend-rule state-based action
// (game/legend_rule_exemption.go, #2177). When the Gallery leaves, the
// next check applies the rule and each controller chooses what to keep.
func init() {
	Register(Spec{
		OracleID:             "7c9c3060-ca9d-4868-84c6-2a95b0fa8885",
		Name:                 "Mirror Gallery",
		Completeness:         CompletenessFull,
		LegendRuleExemptions: LegendRuleDoesntApply(),
	})
}
