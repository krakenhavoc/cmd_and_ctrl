package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Karametra's Blessing — Instant {W} (Theros Beyond Death, S82 OLD, #5514):
//
//	"Target creature gets +2/+2 until end of turn. If it's an enchanted
//	 creature or enchantment creature, it also gains hexproof and
//	 indestructible until end of turn."
//
// The pump is certain; the second sentence is judged as the spell resolves
// (CR 608.2b), after the pump: the target is an enchantment creature or has an
// Aura attached then (effects.Enchanted, a Role included), and only then does
// it gain hexproof and indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d553ae4d-b2a9-47eb-8c2e-ffab8cc31b26",
		Name:         "Karametra's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Karametra's Blessing — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !(c.IsEnchantment() || effects.IsEnchantedByAura(ctx.Game, t.ID)) {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof", "indestructible"},
					Label: "Karametra's Blessing — hexproof and indestructible"}.Apply(ctx)
			}
			return nil
		},
	})
}
