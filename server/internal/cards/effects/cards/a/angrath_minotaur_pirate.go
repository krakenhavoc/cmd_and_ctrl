package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angrath, Minotaur Pirate — Legendary Planeswalker — Angrath {4}{B}{R}, loyalty 5:
//
//	"+2: Angrath deals 1 damage to target opponent or planeswalker and each
//	 creature that player or that planeswalker's controller controls.
//	 −3: Return target Pirate card from your graveyard to the battlefield.
//	 −11: Destroy all creatures target opponent controls. Angrath deals damage
//	 to that player equal to their total power."
//
// +2: one damage event (CR 120.2) to the target and to every creature its player
// controls, the player being the target opponent or the controller of the
// target planeswalker, read as the ability resolves; a target that has left
// does nothing. −3: any card with the Pirate subtype in your graveyard, a
// Kindred one included. −11: the opponent's total power is counted as the
// ability begins to resolve, before the destruction (the creatures that are
// destroyed are the ones whose power is dealt), indestructible ones included in
// the count; creatures are destroyed in one event and the damage follows.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2737a78c-77ed-4bd0-ac0d-39ca2d816e83",
		Name:            "Angrath, Minotaur Pirate",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Angrath deals 1 damage to target opponent or planeswalker and each creature that player or that planeswalker's controller controls.",
				Cost:    effects.LoyaltyCost(2),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Targets: effects.TargetOpponentOrPlaneswalker(),
				Effect:  angrathPlusTwo,
			},
			{
				Label:   "−3: Return target Pirate card from your graveyard to the battlefield.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: effects.ForTargets(effects.Reanimates(0)),
				Targets: effects.TargetCardInGraveyard("target Pirate card in your graveyard", effects.Subtype("Pirate"), effects.YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					_, _ = effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
					return nil
				},
			},
			{
				Label:   "−11: Destroy all creatures target opponent controls. Angrath deals damage to that player equal to their total power.",
				Cost:    effects.LoyaltyCost(-11),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, OpponentsOnly: true}},
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  angrathUltimate,
			},
		},
	})
}

// angrathPlusTwo deals the 1 damage to the target and to each creature its
// player controls, as one damage event.
func angrathPlusTwo(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		owner := t.ID
		if t.Kind == game.TargetCard {
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			owner = c.Controller
		}
		target := t.ID
		return g.DamageInstanceForEffect(func() error {
			if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.Controller == owner
			}), 1)
		})
	}
	return nil
}

// angrathUltimate destroys every creature the target opponent controls and
// then deals damage to them equal to the total power those creatures had.
func angrathUltimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := effects.TargetedPlayerOf(ctx)
	if player == uuid.Nil {
		return nil
	}
	theirs := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Controller == player }
	total := 0
	for _, c := range effects.MatchingBattlefield(ctx, effects.And(effects.Creature(), theirs)) {
		if p := c.CurrentPower(); p > 0 {
			total += p
		}
	}
	return effects.DestroyAllMatching{
		Match:    effects.And(effects.Creature(), theirs),
		ThenTail: effects.WithN(effects.WithCarry(ctx.ItemTail(angrathUltimateDamage), player), total),
	}.Apply(ctx)
}

// angrathUltimateDamage deals the damage once the creatures have been destroyed,
// as many points as the opponent's total power counted before. An on-disk key:
// never renamed, never reused.
var angrathUltimateDamage = effects.CardTail("tail/angrath-minotaur-pirate-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	player := effects.TailCarry0(o)
	if player == uuid.Nil || o.Tail.N <= 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: player, Amount: o.Tail.N}.Apply(ctx)
})
