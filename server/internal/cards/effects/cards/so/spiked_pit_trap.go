package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiked Pit Trap — Artifact {1} (Adventures in the Forgotten Realms,
// #5499):
//
//	"Flash
//	 {5}, {T}, Sacrifice this artifact: Choose target creature, then roll
//	 a d20.
//	 1—9 | This artifact deals 5 damage to that creature.
//	 10—20 | This artifact deals 5 damage to that creature. Create a
//	 Treasure token."
//
// The artifact is sacrificed as a cost, so the damage is dealt by it as it
// last existed (CR 608.2h). If the creature is an illegal target as the
// ability resolves, the ability does nothing, roll included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a0aa1aa1-f011-4c65-9886-95521947beef",
		Name:            "Spiked Pit Trap",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Sacrifice this artifact: Choose target creature, then roll a d20.",
			Purpose: game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.Do(effects.RollDice{Sides: 20, N: 1, Then: spikedPitTrapTable}),
		}},
	})
}

// spikedPitTrapTable is the results table, after the d20.
var spikedPitTrapTable = effects.DiceTail("dice/spiked-pit-trap-table", func(ctx *effects.Context, rolls []int) error {
	if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
		if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 5}).Apply(ctx); err != nil {
			return err
		}
	}
	if rolls[0] >= 10 {
		return effects.TreasureTokens(ctx.Game, ctx.Controller(), 1)
	}
	return nil
})
