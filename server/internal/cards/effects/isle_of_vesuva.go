package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Isle of Vesuva — Plane — Dominaria (Commander Masters, #4870):
//
//	"Whenever a nontoken creature enters, its controller creates a token
//	 that's a copy of that creature.
//	 Whenever chaos ensues, destroy target creature and all other creatures
//	 with the same name as that creature."
//
// A plane (ADR 0216 §6). The first ability triggers for every player's
// nontoken creature, so a token copy it makes does not trigger it again (it
// is a token); the creature's controller makes the copy, from its copiable
// values, or its last-known information if it has already left (CR 707.2).
// The chaos ability (CR 311.7) targets a creature and destroys it and every
// other creature with the same name at once, so if the target is not
// destroyed (it regenerates or is indestructible) the others still are (the
// ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f017380f-97db-4a36-b111-b257aa17ebaf",
		Name:         "Isle of Vesuva",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !c.IsToken()
			}, "Isle of Vesuva — its controller creates a token that's a copy of that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					controller := item.Controller
					if c, ok := g.LookupCardForEffect(id); ok {
						controller = c.Controller
					}
					return CreateTokenCopy{Controller: controller, Copy: id, N: 1}.Apply(NewContext(g, item))
				}),
			Targeting(WheneverChaosEnsues("Isle of Vesuva — destroy target creature and all other creatures with the same name as that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					target, ok := g.LookupCardForEffect(t)
					if !ok {
						return nil
					}
					ids := []uuid.UUID{t}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.InstanceID != t && c.IsCreature() && c.Name == target.Name {
							ids = append(ids, c.InstanceID)
						}
					}
					g.DestroyPermanentsForEffect(ids)
					return nil
				}), TargetCreature("target creature")),
		},
	})
}
