package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rowan, Fearless Sparkmage — Legendary Planeswalker — Rowan {3}{R}{R},
// loyalty 5 (Throne of Eldraine):
//
//	"+1: Up to one target creature gets +3/+0 and gains first strike until
//	     end of turn.
//	 −2: Rowan deals 1 damage to each of up to two target creatures. Those
//	     creatures can't block this turn.
//	 −9: Gain control of all creatures until end of turn. Untap them. They
//	     gain haste until end of turn."
//
// +1 is the pump-and-grant body with an "up to one" target, so none is fine.
// −2 deals the damage to each target still legal and bars those same
// creatures from blocking (Beat a Path's restriction). −9 is Mob Rule's
// steal over every creature, yours included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd83f886-c651-4ad7-82ca-2446c60c8f94",
		Name:            "Rowan, Fearless Sparkmage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature gets +3/+0 and gains first strike until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike")),
				Effect:  effects.PumpAndGrantFirstTarget(3, 0, "Rowan, Fearless Sparkmage — +3/+0 and first strike", "first strike"),
			},
			{
				Label:   "−2: Rowan deals 1 damage to each of up to two target creatures. Those creatures can't block this turn.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  rowanSparkmageDamage,
			},
			{
				Label:   "−9: Gain control of all creatures until end of turn. Untap them. They gain haste until end of turn.",
				Cost:    effects.LoyaltyCost(-9),
				Purpose: game.Purpose{},
				Effect:  rowanSparkmageSteal,
			},
		},
	})
}

// rowanSparkmageDamage is −2: 1 damage to each target still legal, and none
// of those creatures can block this turn.
func rowanSparkmageDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if err := g.DamageInstanceForEffect(func() error {
		for _, id := range ids {
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, id := range ids {
		if err := (effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: "Rowan, Fearless Sparkmage — can't block"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// rowanSparkmageSteal is −9: every creature, yours included.
func rowanSparkmageSteal(g *game.Game, item *game.StackItem) error {
	return effects.RprB28StealAllMatching(effects.NewContext(g, item), "Rowan, Fearless Sparkmage", effects.RprB28CreatureWithPower(func(int) bool { return true }))
}
