package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bouncer's Beatdown — Instant {2}{G}:
//
//	"This spell costs {2} less to cast if it targets a black permanent.
//	 Bouncer's Beatdown deals X damage to target creature or
//	 planeswalker, where X is the greatest power among creatures you
//	 control. If that creature or planeswalker would die this turn,
//	 exile it instead."
//
// The reduction reads the announced target at CR 601.2f (Price of
// Fame's shape, CostsLessIfItTargets). X is not a cost: it is the
// greatest layered power among the caster's creatures as the spell
// resolves (CR 608.2h), and 0 with none. The damage's source is the
// spell, not a creature. The replacement is the spell's (ADR 0108 §1).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "36ac8fc6-98ad-499b-9adf-046433e2c122",
		Name:         "Bouncer's Beatdown",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a black permanent.", OfColor("B")),
		},
		OnResolve: damageFirstTargetExileIfItDies(g2GreatestPowerAmongYourCreatures()),
	})
}
