package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderbolt — Instant {1}{R} (Weatherlight, #5629):
//
//	"Choose one —
//	 • Thunderbolt deals 3 damage to target player or planeswalker.
//	 • Thunderbolt deals 4 damage to target creature with flying."
//
// Two bullets, each with its own target clause (CR 700.2). Glorious
// Decay's second bullet is the flying one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ffae2fa-511a-45bf-93c4-5d53c85324f5",
		Name:         "Thunderbolt",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Thunderbolt deals 3 damage to target player or planeswalker.",
				effects.TargetPlayerOrPlaneswalker(),
				effects.SocB20DamageToTheModeTarget(3)),
			effects.ModeDoing("Thunderbolt deals 4 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				effects.SocB20DamageToTheModeTarget(4)),
		),
	})
}
