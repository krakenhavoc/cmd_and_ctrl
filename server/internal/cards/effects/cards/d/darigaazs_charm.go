package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darigaaz's Charm — Instant {B}{R}{G} (Planeshift, #5615):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Darigaaz's Charm deals 3 damage to any target.
//	 • Target creature gets +3/+3 until end of turn."
//
// A modal spell (CR 700.2) whose bullets are the catalog's mode bodies:
// Kolaghan's Command's graveyard return, its fixed damage at 3, and
// Brigid's Command's +3/+3. Each bullet's target is checked against its
// own clause at announce and again at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ee01801-5e42-4916-a359-64e9975090a7",
		Name:         "Darigaaz's Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTheModesGraveyardTargetToHand),
			effects.ModeWithPurpose(effects.ModeDoing("Darigaaz's Charm deals 3 damage to any target.",
				effects.TargetAny(),
				effects.DealFixedDamageToModesTarget(3)), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+3 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(3, 3, "Darigaaz's Charm")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
		),
	})
}
