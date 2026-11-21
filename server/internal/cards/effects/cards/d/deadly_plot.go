package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deadly Plot — Instant {3}{B} (Foundations, #4465):
//
//	"Choose one —
//	 • Destroy target creature or planeswalker.
//	 • Return target Zombie creature card from your graveyard to the
//	   battlefield tapped."
//
// Molten Collapse's first bullet, and Phoenix Down's tapped return with
// the clause narrowed to a Zombie creature card you own. Each bullet has
// its own target, chosen as the spell is cast (CR 700.2a), and does
// nothing if that target is illegal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6178715c-870d-4710-b758-66e080804ee3",
		Name:         "Deadly Plot",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Return target Zombie creature card from your graveyard to the battlefield tapped.",
				effects.TargetCardInGraveyard("target Zombie creature card from your graveyard", effects.YouOwn(), effects.Creature(), effects.HasSubtype("Zombie")),
				effects.ReturnTheModesTargetToBattlefieldTapped),
		),
	})
}
