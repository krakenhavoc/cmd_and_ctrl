package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scarab of the Unseen — Artifact {2} (Alliances, #5634):
//
//	"{T}, Sacrifice this artifact: Return all Auras attached to target
//	 permanent you own to their owners' hands. Draw a card at the beginning of
//	 the next turn's upkeep."
//
// The target is a permanent its controller OWNS (not merely controls). As the
// ability resolves every Aura attached to it, whoever controls or owns the
// Aura, returns to its owner's hand (CR 400.3), and the delayed draw at the
// next upkeep at the table is scheduled either way (CR 603.7a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f085af7-ce66-4019-9516-d1650ee7b6e0",
		Name:         "Scarab of the Unseen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this artifact: Return all Auras attached to target permanent you own to their owners' hands. Draw a card at the beginning of the next turn's upkeep.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target permanent you own", effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if t, ok := effects.FirstLegalTarget(ctx); ok {
					var auras []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsAttachedTo(t.ID) && c.IsAura() {
							auras = append(auras, c.InstanceID)
						}
					}
					for _, id := range auras {
						if err := (effects.BounceToHand{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
				}
				return effects.DrawACardAtTheNextUpkeep("Scarab of the Unseen — draw a card").Apply(ctx)
			},
		}},
	})
}
