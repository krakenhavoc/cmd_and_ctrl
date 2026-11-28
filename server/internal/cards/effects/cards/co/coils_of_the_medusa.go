package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coils of the Medusa — Enchantment — Aura {1}{B} (Weatherlight):
//
//	"Enchant creature
//	 Enchanted creature gets +1/-1.
//	 Sacrifice this Aura: Destroy all non-Wall creatures blocking enchanted
//	 creature."
//
// "Enchanted creature" is the creature the Aura was attached to as it was
// sacrificed (CR 608.2h); the blockers are read as the ability resolves, and
// those that are Walls are left alone. A creature that is not attacking has no
// blockers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02a8a4e6-2081-4903-81ad-96e5a949dff8",
		Name:         "Coils of the Medusa",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, -1)},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this Aura: Destroy all non-Wall creatures blocking enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				host, ok := effects.EnchantedCreatureAsItResolves(ctx)
				if !ok {
					return nil
				}
				var ids []uuid.UUID
				for _, id := range effects.BlockersOf(g, host) {
					if c, ok := g.LookupCardForEffect(id); ok && !c.HasSubtype(ctx.TextSubtype("Wall")) {
						ids = append(ids, id)
					}
				}
				if len(ids) > 0 {
					g.DestroyPermanentsForEffect(ids)
				}
				return nil
			},
		}},
	})
}
