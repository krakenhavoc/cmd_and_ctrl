package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4da9b548-898d-431b-8aa5-9cb7bf579219",
		Name:         "Set Phasers to . . .",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Stun — Tap target creature and put a stun counter on it.",
				effects.TargetCreature("target creature"), effects.TrekB9TapAndStunTheModesTarget),
			effects.ModeDoing("Kill — This spell deals 3 damage to target attacking creature.",
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
				effects.DealFixedDamageToModesTarget(3)),
		),
	})
}
