package effects

// Mortify — Instant {1}{W}{B}:
//
//	"Destroy target creature or enchantment. It can't be regenerated."
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c):
// the rider rides the destroy route onto the CR 614 event and the
// regeneration built-in declines. The shield is not spent
// (CR 701.19c).
func init() {
	Register(Spec{
		OracleID:     "faa01ed1-ccfa-4e58-951f-cd81f9068027",
		Name:         "Mortify",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or enchantment", Or(Creature(), Enchantment())),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
