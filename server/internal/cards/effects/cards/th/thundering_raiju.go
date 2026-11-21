package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Raiju — Creature — Spirit {2}{R}{R}, 3/3 (Kamigawa: Neon
// Dynasty):
//
//	"Haste
//	 Whenever this creature attacks, put a +1/+1 counter on target creature
//	 you control. Then this creature deals X damage to each opponent, where
//	 X is the number of modified creatures you control other than this
//	 creature. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// X is counted after the counter lands ("then"), so the counter just put on
// another creature makes it modified and counts; one put on the Raiju itself
// makes only the Raiju modified, which is excluded. If the counter's target
// is gone the damage still happens (CR 608.2b: only the target is illegal).
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Thundering Raiju — put a +1/+1 counter on target creature you control, then deal damage to each opponent",
		func(g *game.Game, item *game.StackItem) error {
			if err := effects.PutPlusOneCounterOnEachLegalTarget(g, item); err != nil {
				return err
			}
			x := effects.NeoB10ModifiedCreatureCount(g, item.Controller, item.SourceCardID)
			if x <= 0 {
				return nil
			}
			return effects.DamageToEachOpponent(g, item, x)
		})
	attack.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:        "3d713c11-36a7-4b19-8c42-a9b61ef5902c",
		Name:            "Thundering Raiju",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{attack},
	})
}
