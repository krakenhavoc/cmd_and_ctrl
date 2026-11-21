package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unfinished Business — Sorcery {3}{W}{W} (#3656):
//
//	"Return target creature card from your graveyard to the
//	 battlefield, then return up to two target Aura and/or Equipment
//	 cards from your graveyard to the battlefield attached to that
//	 creature. (If the Auras can't enchant that creature, they remain
//	 in your graveyard.)"
//
// Mantle of the Ancients' return-attached (ReturnFromGraveyardTogether
// with AttachedTo) as the second half of a two-clause spell, with the
// first clause a plain reanimation. An Aura whose enchant clause the
// creature doesn't meet stays in the graveyard (CR 303.4j,
// canBeAttachedTo); an Equipment needs only a creature. If the creature
// can't be returned (it left the graveyard in response), nothing is
// attached to, so the second half returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "800cfd70-6a97-40eb-8dc0-40b762e4db76",
		Name:         "Unfinished Business",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn()),
			effects.TargetCardInGraveyard("up to two target Aura and/or Equipment cards in your graveyard",
				effects.YouOwn(), effects.Or(effects.HasSubtype("Aura"), effects.HasSubtype("Equipment"))).WithCount(0, 2),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			creature, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.ReturnFromGraveyard{Target: creature.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
			if !g.Battlefield.Contains(creature.ID) {
				return nil
			}
			var ids []uuid.UUID
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind != game.TargetCard {
					continue
				}
				card, ok := g.LookupCardForEffect(t.ID)
				if ok && effects.CanBeAttachedTo(g, item.Controller, card, creature.ID) {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{
				Targets:    ids,
				AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: creature.ID},
			}.Apply(ctx)
		},
	})
}
