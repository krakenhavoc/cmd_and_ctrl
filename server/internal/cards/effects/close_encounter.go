package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Close Encounter — Instant {1}{G} (Edge of Eternities, #4461):
//
//	"As an additional cost to cast this spell, choose a creature you
//	 control or a warped creature card you own in exile.
//	 Close Encounter deals damage equal to the power of the chosen
//	 creature or card to target creature."
//
// The additional cost is ChooseCreatureOrWarpedCardCost: the caster names
// one creature they control or one warped creature card they own in exile
// (a card exiled by warp's delayed trigger and still there, CR 702.185b),
// on reveal_ids; with neither the spell can't be cast. Nothing is moved,
// revealed or tapped. The chosen object rides the spell's payment record,
// and the damage is its power read at resolution: a creature that is still
// on the battlefield counts as it is now, one that has left counts as it
// last existed there (CR 608.2h, the rulings), and a card in exile counts its
// printed power. The damage is Close Encounter's own, to the creature
// targeted as the spell was cast; if that target is gone the spell fizzles.
//
// Seam: RevealCost.ChooseCreatureOrWarped and CastPermission.Warped.
func init() {
	Register(Spec{
		OracleID:       "032fdd08-2ea7-442e-8162-0a8e5a585253",
		Name:           "Close Encounter",
		Completeness:   CompletenessFull,
		AdditionalCost: ChooseCreatureOrWarpedCardCost(),
		Targets:        TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			power, ok := ctx.TappedPower()
			if !ok || power <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: power}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
