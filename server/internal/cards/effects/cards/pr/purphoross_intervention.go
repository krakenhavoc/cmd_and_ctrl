package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Purphoros's Intervention — Sorcery {X}{R} (Theros Beyond Death, #5514):
//
//	"Choose one —
//	 • Create an X/1 red Elemental creature token with trample and haste.
//	   Sacrifice it at the beginning of the next end step.
//	 • Purphoros's Intervention deals twice X damage to target creature or
//	   planeswalker."
//
// X is the announced X (CR 107.3). The token is sized as it is created
// (PurphorossInterventionElemental) and sacrificed at the beginning of the next
// end step (CR 603.7); the second mode deals 2X from the spell to its one target,
// chosen as the spell is cast (CR 700.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e190f086-ebee-447f-907f-7b37ec8824ac",
		Name:         "Purphoros's Intervention",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Create an X/1 red Elemental creature token with trample and haste. Sacrifice it at the beginning of the next end step.",
				nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					made, err := effects.PurphorossInterventionElemental(ctx, ctx.X())
					if err != nil {
						return err
					}
					return effects.S82OneB12SacrificeAtEnd(ctx, "Purphoros's Intervention — sacrifice the Elemental", made)
				}),
			effects.ModeDoing("Purphoros's Intervention deals twice X damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2 * ctx.X()}.Apply(ctx)
				}),
		),
	})
}
