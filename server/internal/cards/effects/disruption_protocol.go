package effects

// Disruption Protocol — Instant {U}{U} (#3868):
//
//	"As an additional cost to cast this spell, tap an untapped artifact
//	 you control or pay {1}.
//	 Counter target spell."
//
// The either/or cost (ADR 0100 §2), branches in printed order: branch 0
// taps an untapped artifact the caster controls (TapUntappedCost, the
// artifact named on cast_spell as teamwork_ids), branch 1 pays {1}. The
// artifact tapped to pay is not tapped for the {U}{U} (CR 601.2g before
// 601.2h), and with no untapped artifact only the {1} branch is offered
// (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ca3306b-06f7-4a09-b904-f5540d11ddf1",
		Name:         "Disruption Protocol",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			TapUntappedCost(1, "an untapped artifact you control", Artifact()).Keyed("tap"),
			ManaAdditionalCost("{1}").Keyed("mana"),
		),
		Targets:   TargetSpell("target spell"),
		OnResolve: counterTheTargetSpell,
	})
}
