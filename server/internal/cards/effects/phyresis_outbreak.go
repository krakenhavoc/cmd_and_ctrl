package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyresis Outbreak — Sorcery {2}{B}: "Each opponent gets a poison counter.
// Then each creature your opponents control gets -1/-1 until end of turn for
// each poison counter its controller has."
//
// The set of creatures and each one's amount are fixed as the spell resolves,
// after the counters are given (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47148b5d-3850-4614-822e-6ef2755996f8",
		Name:         "Phyresis Outbreak",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, AmountIsX: true, OpponentsOnly: true}},
		OnResolve: ResolveSteps(
			s82OneB04PoisonStep(),
			Instruction(func(ctx *Context) error {
				for _, opp := range ctx.Opponents() {
					p := ctx.Game.PlayerByIDForEffect(opp)
					if p == nil || p.Counters[game.CounterPoison] == 0 {
						continue
					}
					n := p.Counters[game.CounterPoison]
					owner := opp
					err := BoostUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return c.IsCreature() && c.Controller == owner
						},
						Power: -n, Toughness: -n,
						Label: "Phyresis Outbreak — -1/-1 for each poison counter",
					}.Apply(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),
		),
	})
}
