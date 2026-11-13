package effects

// Opposition — Enchantment {2}{U}{U} (Urza's Destiny, #5621):
//
//	"Tap an untapped creature you control: Tap target artifact,
//	 creature, or land."
//
// The cost is tapping a creature, not the {T} symbol (CR 302.6), so a
// creature that arrived this turn can pay it, once per untapped
// creature. The target is any player's artifact, creature or land; one
// gone by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf0b252d-1295-44e4-bac3-113b5732a2a9",
		Name:         "Opposition",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Tap target artifact, creature, or land.",
			Cost:    udsB2TapAnUntappedCreatureYouControl(),
			Targets: TargetPermanent("target artifact, creature, or land", Or(Artifact(), Creature(), Land())),
			Effect:  udsB2TapTheTarget,
		}},
	})
}
