package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// brambleguardCaptainLabel is the trigger's key: an on-disk identity.
const brambleguardCaptainLabel = "Brambleguard Captain — target creature you control gets +X/+0"

// Brambleguard Captain — Creature — Mouse Soldier {3}{R}, 2/3
// (Bloomburrow, #4467):
//
//	"At the beginning of combat on your turn, target creature you
//	 control gets +X/+0 until end of turn, where X is this creature's
//	 power."
//
// Gruul Beastmaster's body (gruulBeastmasterPump's shape) on a
// beginning-of-combat trigger. X is the Captain's power as the trigger
// resolves, or its last-known power if it has left (CR 608.2h); the
// Captain may target itself, since the clause does not say "another".
// A power of 0 or less gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "104baa4e-a3f3-45dc-979d-4c58b659ba1a",
		Name:         "Brambleguard Captain",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat(brambleguardCaptainLabel, brambleguardCaptainPump),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

func brambleguardCaptainPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return vsPumpTheTarget(ctx, info.Power, 0, brambleguardCaptainLabel)
}
