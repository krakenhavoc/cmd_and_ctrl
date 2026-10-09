package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Martyr of Ashes — Creature — Human Shaman {R}, 1/1:
//
//	"{2}, Reveal X red cards from your hand, Sacrifice this creature:
//	 This creature deals X damage to each creature without flying."
//
// The reveal cost is effects.RevealX (#2598, ADR 0020's 2026-10-08
// amendment): its count is the X announced with the activation, the
// revealed cards stay in the hand, and the mana cost is {2} whatever X
// is. The sacrifice is a cost, so by resolution the Martyr is gone and
// is still the damage's source as it last existed (CR 113.7a, 608.2h,
// ctx.SourceRef). The creatures are snapshotted before the first point
// lands and "without flying" is read after layers, so a granted flying
// keeps a creature out, exactly as Earthquake does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f00bc45-7c65-4455-9db0-58bd79bcdb4b",
		Name:         "Martyr of Ashes",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, Partial: true}},
		Activated: []ActivatedAbility{{
			Label: "{2}, Reveal X red cards from your hand, Sacrifice this creature: This creature deals X damage to each creature without flying.",
			Cost:  Plus(ManaCost("{2}"), RevealX("X red cards", "R"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := ctx.X()
				if x <= 0 {
					return nil
				}
				ref, hasRef := ctx.SourceRef()
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, c := range MatchingBattlefield(ctx, And(Creature(), WithoutKeyword("flying"))) {
						d := DealDamage{Source: ctx.Source(), Target: c.InstanceID, Amount: x}
						if hasRef {
							d = DealDamage{SourceObject: &ref, Target: c.InstanceID, Amount: x}
						}
						if err := d.Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}
