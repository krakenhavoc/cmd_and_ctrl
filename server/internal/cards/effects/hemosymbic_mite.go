package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hemosymbic Mite — Creature — Mite {G}, 1/1 (Edge of Eternities, #4461):
//
//	"Whenever this creature becomes tapped, another target creature you
//	 control gets +X/+X until end of turn, where X is this creature's
//	 power."
//
// Attentive Sunscribe's "becomes tapped" (ThisBecameTapped: an attack
// without vigilance, a cost, any other tap; CR 701.26a) with Ashroot
// Animist's +X/+X on another target creature you control. X is the
// Mite's power as the ability resolves, or its last-known power if it has
// left (sourcePowerNow, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "103e0b63-fbde-4054-91f7-dd04b40b59ef",
		Name:         "Hemosymbic Mite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Hemosymbic Mite — another target creature you control gets +X/+X",
				Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
					x := sourcePowerNow(ctx)
					return BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Hemosymbic Mite — +X/+X"}.Apply(ctx)
				}))),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
