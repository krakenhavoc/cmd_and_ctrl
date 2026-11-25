package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pestilent Haze — Sorcery {1}{B}{B} (Core Set 2021, #5511):
//
//	"Choose one —
//	 • All creatures get -2/-2 until end of turn.
//	 • Remove two loyalty counters from each planeswalker."
//
// Mode 0 is Biting Rain's -2/-2 on the creatures there are as it resolves
// (CR 611.2c). Mode 1 removes two loyalty counters from every planeswalker
// (a planeswalker with fewer loses what it has; at 0 loyalty it is put into
// its owner's graveyard as a state-based action, CR 704.5i).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b24d1540-ed8b-4df3-abfe-1b8492a58388",
		Name:         "Pestilent Haze",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Modes: effects.ChooseOne(
			effects.Mode("All creatures get -2/-2 until end of turn."),
			effects.Mode("Remove two loyalty counters from each planeswalker."),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(0) {
					return nil
				}
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Pestilent Haze — -2/-2 until end of turn"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.HasMode(1) {
					return nil
				}
				var ids []game.Card
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.HasCardType("planeswalker") {
						ids = append(ids, c)
					}
				}
				for _, c := range ids {
					n := min(2, c.Counters[game.CounterLoyalty])
					if n <= 0 {
						continue
					}
					if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterLoyalty, N: -n}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
