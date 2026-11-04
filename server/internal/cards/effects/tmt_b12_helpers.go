package effects

// tmt_b12_helpers.go — Teenage Mutant Ninja Turtles, batch 12 (#4457):
// the pieces the batch's sneak cards share.

// sneakAltCostKey is the key Sneak(cost) gives its offer (sneak.go), the
// value a cast record's AltCost holds when the sneak cost was paid.
const sneakAltCostKey = "sneak"

// sneakCostWasPaid reports whether a permanent's cast record says its
// sneak cost was paid (CR 702.190): "if this creature's sneak cost was
// paid", read off ObjectSnapshot.AltCost the way Adipose Offspring reads
// emerge. A resolving spell asks Context.PaidAltCost(sneakAltCostKey).
func sneakCostWasPaid(altCost string) bool {
	return altCost == sneakAltCostKey
}
