package effects

// Iron Will — Instant {W} (Urza's Legacy, #5623):
//
//	"Target creature gets +0/+4 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Lightshield Parry's shape at +0/+4: a toughness boost that ends in the
// cleanup step (CR 514.2), with cycling (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc0decb9-32da-47aa-a51a-0d9c623c534a",
		Name:         "Iron Will",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 4)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(0, 4, "Iron Will — +0/+4"),
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
