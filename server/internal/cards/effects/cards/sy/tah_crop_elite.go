package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tah-Crop Elite — Creature — Bird Warrior {3}{W}, 2/2:
//
//	"Flying
//	 You may exert this creature as it attacks. When you do, creatures
//	 you control get +1/+1 until end of turn. (An exerted creature
//	 won't untap during your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h). The set of creatures is locked as the trigger
// resolves (CR 611.2c), Tah-Crop Elite included, so a creature that
// arrives later this turn gets nothing.
//
// No simplification.
func init() {
	const label = "Tah-Crop Elite — creatures you control get +1/+1 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "9af1750e-90cc-446a-9ba2-d44e03c9d01b",
		Name:            "Tah-Crop Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ExertOnAttack:   effects.ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			effects.WhenExerted(label, func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
