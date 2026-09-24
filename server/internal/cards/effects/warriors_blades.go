package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warrior's Blades — Artifact — Equipment for {2}{R}{W}:
//
//	"When this Equipment enters, it deals 3 damage to any target and
//	 you gain 3 life.
//	 Equipped creature gets +2/+1.
//	 Equip {3}. This ability costs {1} less to activate for each +1/+1
//	 counter on the creature it targets."
//
// A proof card for the ability's own cost clause (#1296,
// ActivatedAbility.CostModifiers) keyed to the TARGET: the +1/+1
// counters counted are the ones on the creature the equip announces,
// as they stand when the cost is determined (CR 602.2b → 601.2c, then
// 601.2f). Three or more counters equip for {0}; the reduction never
// goes below zero generic (the engine's floor).
//
// The ETB is one targeted trigger: the damage and the life are one
// ability, so a target that becomes illegal in response counters the
// whole thing (CR 608.2b) and no life is gained.
func init() {
	equip := EquipAbility("{3}")
	equip.CostModifiers = []game.CostModifier{
		CostsLessForTheCardItTargets(
			"This ability costs {1} less to activate for each +1/+1 counter on the creature it targets.",
			CountersOf(game.CounterPlusOne)),
	}
	Register(Spec{
		OracleID:     "f47419a5-b975-4934-a987-ea064a0c7c1a",
		Name:         "Warrior's Blades",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Warrior's Blades — 3 damage to any target and you gain 3 life", warriorsBladesETB),
				TargetAny()),
		},
		Static:    []game.StaticAbility{PumpAttached(2, 1)},
		Activated: []ActivatedAbility{equip},
	})
}

// warriorsBladesETB deals the 3 damage and gains the 3 life. The
// damage source is the Equipment ("it deals"), read off the item.
func warriorsBladesETB(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
			return err
		}
	}
	return GainLife{Amount: 3}.Apply(ctx)
}
