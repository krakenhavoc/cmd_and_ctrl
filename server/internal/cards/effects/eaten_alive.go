package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eaten Alive — Sorcery {B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature or
//	 pay {3}{B}. Exile target creature or planeswalker."
//
// The either/or cost (ADR 0100 §2); the {3}{B} branch joins the total at
// CR 601.2f through the one pricer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d437ecc2-2fd3-4ad3-b23e-217f55e58dae",
		Name:         "Eaten Alive",
		Completeness: CompletenessFull,
		AdditionalCost: EitherCost(
			SacrificeCost("a creature", Creature()).Keyed("sacrifice"),
			ManaAdditionalCost("{3}{B}").Keyed("mana"),
		),
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return exileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
