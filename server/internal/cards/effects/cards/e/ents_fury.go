package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ent's Fury — Sorcery {1}{G} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Put a +1/+1 counter on target creature you control if its power is 4 or
//	 greater. Then that creature gets +1/+1 until end of turn and fights target
//	 creature you don't control."
//
// Two target clauses. The counter is put only if the creature's power is 4 or
// more as the spell resolves (through the layers). Then the creature, if its
// target is still legal, gets +1/+1 until end of turn and fights the second
// target (Amok Time's slot-zero-fights-slot-one, CR 701.14b: nothing fights
// unless both are still legal). The counter is put first, so a 3-power creature
// is a 4/5 for the fight only through the +1/+1, and a creature that reached 4
// power gets the counter and the pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bc3c7fd-e371-467f-973c-4c9f0d1ba0bc",
		Name:         "Ent's Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || c.CurrentPower() < 4 {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Ent's Fury — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.SlotZeroFightsSlotOne(ctx.Game, ctx.Item)
			}),
		),
	})
}
