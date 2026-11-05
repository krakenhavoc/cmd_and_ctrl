package effects

// Set Phasers to . . . — Instant {W} (Star Trek, #4973):
//
//	"Choose one —
//	 • Stun — Tap target creature and put a stun counter on it.
//	 • Kill — This spell deals 3 damage to target attacking creature."
//
// Stun is Cryotheory Adept's tap-and-stun on the mode's target (the stun
// counter is CR 122.1d's untap replacement, which the engine applies);
// Kill is fixed damage to an attacking creature, legal only during the
// combat that declared it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4da9b548-898d-431b-8aa5-9cb7bf579219",
		Name:         "Set Phasers to . . .",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Stun — Tap target creature and put a stun counter on it.",
				TargetCreature("target creature"), trekB9TapAndStunTheModesTarget),
			ModeDoing("Kill — This spell deals 3 damage to target attacking creature.",
				TargetCreature("target attacking creature", AttackingCreature()),
				DealFixedDamageToModesTarget(3)),
		),
	})
}
