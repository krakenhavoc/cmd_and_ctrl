package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra, Bold Pyromancer — Legendary Planeswalker — Chandra {4}{R}{R},
// loyalty 5 (Dominaria, S82 OLD batch 215, #5530):
//
//	"+1: Add {R}{R}. Chandra deals 2 damage to target player.
//	 −3: Chandra deals 3 damage to target creature or planeswalker.
//	 −7: Chandra deals 10 damage to target player and each creature and
//	 planeswalker they control."
//
// The +1 adds the mana as it resolves (CR 106.4: it empties at the end of the
// step) and then deals the damage to the one target, which is re-checked as it
// resolves (CR 608.2b): with the player gone the mana is not added either, as
// the whole ability is removed. The −3 is a plain damage to one creature or
// planeswalker. The −7 reads the target player's permanents as it resolves and
// deals all the damage in one damage instance (CR 120.2): the player, then
// each creature and planeswalker that player controls, Chandra as the source
// (CR 608.2h, so the damage is dealt even when she has left).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1cdba3c1-19e1-43c3-a0d8-5fb2b88060d8",
		Name:            "Chandra, Bold Pyromancer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Add {R}{R}. Chandra deals 2 damage to target player.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target player"),
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2))},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddMana{Player: item.Controller, Produced: "{R}{R}"}).Apply(ctx); err != nil {
						return err
					}
					return effects.DealDamageToTheTarget(2)(g, item)
				},
			},
			{
				Label:   "−3: Chandra deals 3 damage to target creature or planeswalker.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3))},
				Effect:  effects.DealDamageToTheTarget(3),
			},
			{
				Label:   "−7: Chandra deals 10 damage to target player and each creature and planeswalker they control.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPlayer("target player"),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 10, OpponentsOnly: true}},
				Effect:  chandraBoldPyromancerBlast,
			},
		},
	})
}

// chandraBoldPyromancerBlast is the −7: ten damage, in one instance, to the
// target player and every creature and planeswalker they control.
func chandraBoldPyromancerBlast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var player uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			player = t.ID
		}
	}
	if player == uuid.Nil {
		return nil
	}
	var victims []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && (c.IsCreature() || c.IsPlaneswalker()) {
			victims = append(victims, c.InstanceID)
		}
	}
	return g.DamageInstanceForEffect(func() error {
		if err := (effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 10}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		for _, id := range victims {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 10}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return nil
	})
}
