package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snakeform — Instant {2}{G/U} (Ravnica: Clue Edition, #5929):
//
//	"Until end of turn, target creature loses all abilities and becomes a
//	 green Snake with base power and toughness 1/1.
//	 Draw a card."
//
// Fractalize's one continuous effect until end of turn: layer 4 the creature
// type becomes Snake, layer 5 the colour is set to green, layer 6 all
// abilities are lost and layer 7b the base power and toughness are 1/1
// (counters and pumps still apply on top). Then the controller draws. A
// target that became illegal counters the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76ae07ce-ca0c-4308-b8bd-d999eea0a65f",
		Name:         "Snakeform",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return untilEndOfTurn(ctx, id, nil,
					"Snakeform — loses all abilities and becomes a green Snake with base power and toughness 1/1 until end of turn",
					game.SetCreatureTypesMod(ctx.TextSubtypes("Snake")...),
					game.SetColorsMod(ctx.TextColors("G")...),
					game.LoseAllAbilitiesMod(),
					game.SetBasePowerMod(1), game.SetBaseToughnessMod(1))
			}),
			Instruction(func(ctx *Context) error { return DrawCards{N: 1}.Apply(ctx) }),
		),
	})
}
