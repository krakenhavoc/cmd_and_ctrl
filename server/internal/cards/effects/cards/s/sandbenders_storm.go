package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandbenders' Storm — Instant {3}{W} (Avatar: The Last Airbender, #4459):
//
//	"Choose one —
//	 • Destroy target creature with power 4 or greater.
//	 • Earthbend 3. (Target land you control becomes a 0/0 creature with
//	   haste that's still a land. Put three +1/+1 counters on it. When it
//	   dies or is exiled, return it to the battlefield tapped.)"
//
// The first bullet is Collective Effort's destroy bullet: power is read as
// the target is chosen and again as the spell resolves (CR 608.2b), so a
// creature shrunk below 4 in response is no longer a legal target and is
// not destroyed. The second is earthbend (CR 701.66a) on the bullet's own
// target land, Dai Li Indoctrination's bullet with three counters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfc8b8de-be1d-4f7f-bb21-4761a0e7db4e",
		Name:         "Sandbenders' Storm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Earthbend 3.", effects.EarthbendTargets(), effects.EarthbendTheModesTarget(3)),
		),
	})
}
