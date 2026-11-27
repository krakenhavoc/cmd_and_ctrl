package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthcraft — Enchantment {1}{G} (Tempest, S82 OLD b753, #5628):
//
//	"Tap an untapped creature you control: Untap target basic land."
//
// Bramblesnap's cost (TapAnUntapped: not the {T} symbol, so a creature that
// came under your control this turn may pay) and Murasa Rootgrazer's basic
// land target. The land is any basic land, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50aa7aff-1f01-4224-9a83-01f74d703ec2",
		Name:         "Earthcraft",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped creature you control: Untap target basic land.",
			Cost:    effects.TapAnUntapped("an untapped creature you control", effects.Creature()),
			Targets: effects.TargetPermanent("target basic land", effects.OldB096BasicLandMatch),
			Effect:  effects.UntapTheTarget,
		}},
	})
}
