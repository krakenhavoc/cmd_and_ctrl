package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bogslither's Embrace — Sorcery {1}{B}:
//
//	"As an additional cost to cast this spell, blight 1 or pay {3}. (To
//	 blight 1, put a -1/-1 counter on a creature you control.) Exile
//	 target creature."
//
// The either/or cost (ADR 0100 §2). The blight branch is the #1703 cost
// component in a branch: the caster names the creature on blight_ids,
// the counter goes through the CR 614 window as a cost, and a caster
// with no creature cannot choose it (CR 701.68b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21695c70-f746-4d77-b6d3-e2713a5c929e",
		Name:         "Bogslither's Embrace",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			BlightCost(1).Keyed("blight"),
			ManaAdditionalCost("{3}").Keyed("mana"),
		),
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
