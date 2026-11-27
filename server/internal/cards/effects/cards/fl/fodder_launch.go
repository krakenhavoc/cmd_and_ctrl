package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fodder Launch — Kindred Sorcery — Goblin {3}{B} (Lorwyn):
//
//	"As an additional cost to cast this spell, sacrifice a Goblin.
//	 Target creature gets -5/-5 until end of turn. Fodder Launch deals 5 damage
//	 to that creature's controller."
//
// The Goblin is paid on casting (CR 601.2h), before the target is read. The
// creature's controller is read as the spell resolves, before the shrink can
// kill it (CR 608.2h); the shrink comes first, then the damage, in printed
// order. A target that is illegal on resolution counters the whole spell
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bec0127a-ab3c-4923-93c2-c43ba8998091",
		Name:           "Fodder Launch",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a Goblin", effects.Subtype("Goblin")),
		Purpose:        game.Purpose{Targets: game.ForTargets(effects.HarmUntilEOT(0, -5, -5))},
		Targets:        effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			target, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			controller := target.Controller
			if err := effects.BoostTheTargetUntilEOT(-5, -5, "Fodder Launch — -5/-5")(item, ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 5}.Apply(ctx)
		},
	})
}
