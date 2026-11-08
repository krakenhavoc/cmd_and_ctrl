package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rat Out — Instant {B} (Wilds of Eldraine, #4471):
//
//	"Up to one target creature gets -1/-1 until end of turn. You create a
//	 1/1 black Rat creature token with "This token can't block.""
//
// The shrink is Dissection Practice's "up to one" -1/-1 clause; the Rat is
// the catalog's can't-block Rat. With no creature chosen the spell still
// resolves and makes the Rat. A chosen creature that is an illegal target
// as the spell resolves leaves the spell with no legal target, so it does
// nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37043e1e-4556-45f7-9484-15dd8ed5272b",
		Name:         "Rat Out",
		Completeness: CompletenessFull,
		Targets:      UpToOneTargetCreature("up to one target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -1, -1)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if target := FirstLegalBattlefieldTarget(ctx); target != uuid.Nil {
				if err := (BoostUntilEOT{Target: target, Power: -1, Toughness: -1, Label: "Rat Out — -1/-1 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return CreateToken{Controller: ctx.Controller(), Template: CantBlockRatToken(), N: 1}.Apply(ctx)
		},
	})
}
