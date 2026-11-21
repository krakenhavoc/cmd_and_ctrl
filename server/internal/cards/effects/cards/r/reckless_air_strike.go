package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reckless Air Strike — Sorcery {R} (Core Set 2020, #5518):
//
//	"Choose one —
//	 • Reckless Air Strike deals 3 damage to target creature with flying.
//	 • Destroy target artifact."
//
// Thunderbolt's flying bullet and Qutrub Forayer's destroy bullet, each on
// its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89247308-e16f-4f60-a2a2-0b5e297c1fc1",
		Name:         "Reckless Air Strike",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Reckless Air Strike deals 3 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.SocB20DamageToTheModeTarget(3)),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
		),
	})
}
