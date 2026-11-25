package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ixalan's Binding — Enchantment {3}{W} (Ixalan, S82 older sets):
//
//	"When this enchantment enters, exile target nonland permanent an opponent
//	 controls until this enchantment leaves the battlefield.
//	 Your opponents can't cast spells with the same name as the exiled card."
//
// Aang's Iceberg's exile-until-leaves (CR 610.3: the card returns when the
// enchantment leaves, and a trigger that has lost its target does not
// resolve) with the CR 607.2a link stamped on the card (ExileUntil.Linked),
// and Ashiok's Erasure's cast restriction read live off the card
// exiled with this object: each opponent of its controller can't cast a spell
// whose name is the exiled card's, and it lifts the moment the enchantment is
// gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a2d0df9-950e-4c3d-8a08-275f179e33c7",
		Name:         "Ixalan's Binding",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ixalan's Binding — exile target nonland permanent an opponent controls until this enchantment leaves the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.ExileUntil{Target: t.ID, ThisLeaves: true, Linked: true,
							Label: "Ixalan's Binding — the exiled card returns when Ixalan's Binding leaves the battlefield"}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
		CastRestrictions: []game.CastRestriction{effects.OpponentsCantCastSpellsNamedLikeTheExiledCard()},
	})
}
