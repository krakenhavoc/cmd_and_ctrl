package effects

// Putrefy — Instant {1}{B}{G}:
//
//	"Destroy target artifact or creature. It can't be regenerated."
//
// Mortify's Golgari counterpart — artifact instead of enchantment.
// Same regeneration rider, enforced since #667.
func init() {
	Register(Spec{
		OracleID:     "9b271430-f53d-42d6-a547-2f286dd9bcb6",
		Name:         "Putrefy",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve:    destroyTheTargetPermanentNoRegen,
	})
}
