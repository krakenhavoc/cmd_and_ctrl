package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// White Sun's Twilight — Sorcery {X}{W}{W}:
//
//	"You gain X life. Create X 1/1 colorless Phyrexian Mite artifact
//	 creature tokens with toxic 1 and 'This token can't block.' If X is
//	 5 or more, destroy all other creatures. (Players dealt combat
//	 damage by a creature with toxic 1 also get a poison counter.)"
//
// "All other creatures" is every creature except the Mites this very
// spell just created: the destroy rides the token creation's
// continuation and spares exactly the tokens that landed (so a
// Doubling Season's extra Mites are spared too, and a Mite that was
// already on the table is not).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a38828be-781e-4340-9c6f-f40b1d34773f",
		Name:         "White Sun's Twilight",
		Completeness: CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: ResolveSteps(
			// The Mites wait for a prompt the life gain asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				return GainLife{Amount: ctx.X()}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				x := ctx.X()
				return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: ctx.Item.Controller,
					Source:     ctx.Source(),
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(PhyrexianMiteToken()), Count: x}},
				}, ctx.TailN(whiteSunsTwilightSweep, x))
			}),
		),
	})
}

// whiteSunsTwilightSweep is "if X is 5 or more, destroy all other creatures"
// as a keyed tail (#4253 slice 6): X is the tail's N, the Mites that were
// created are the ones spared. An on-disk key.
var whiteSunsTwilightSweep = CardTail("white-suns-twilight/destroy-all-other-creatures", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N < 5 {
		return nil
	}
	spared := make(map[uuid.UUID]bool, len(o.IDs))
	for _, id := range o.IDs {
		spared[id] = true
	}
	return DestroyAllMatching{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && !spared[c.InstanceID]
	}}.Apply(ctx)
})
