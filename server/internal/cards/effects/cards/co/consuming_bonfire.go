package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Consuming Bonfire — Kindred Sorcery — Elemental {3}{R}{R} (Lorwyn):
//
//	"Choose one —
//	 • Consuming Bonfire deals 4 damage to target non-Elemental creature.
//	 • Consuming Bonfire deals 7 damage to target Treefolk creature."
//
// Barbed Lightning's two bullets, each with its own target clause and its
// own damage, the bullet's target re-checked on resolution (CR 608.2b). The
// Elemental and Treefolk words are read through the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1feea85-2f06-43aa-9b3c-eb63085276f5",
		Name:         "Consuming Bonfire",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Consuming Bonfire deals 4 damage to target non-Elemental creature.",
				effects.TargetCreature("target non-Elemental creature", effects.Not(effects.HasSubtype("Elemental"))),
				consumingBonfireDamage(4)), effects.ForTargets(effects.DamageToTarget(0, 4))),
			effects.ModeWithPurpose(effects.ModeDoing("Consuming Bonfire deals 7 damage to target Treefolk creature.",
				effects.TargetCreature("target Treefolk creature", effects.HasSubtype("Treefolk")),
				consumingBonfireDamage(7)), effects.ForTargets(effects.DamageToTarget(0, 7))),
		),
	})
}

// consumingBonfireDamage is either bullet: n damage to the bullet's own target,
// if it is still legal.
func consumingBonfireDamage(n int) func(*game.StackItem, *effects.Context, int) error {
	return effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
		return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
	})
}
