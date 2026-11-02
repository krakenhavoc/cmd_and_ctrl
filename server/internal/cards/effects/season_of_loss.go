package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Season of Loss — Sorcery {3}{B}{B} (Bloomburrow, #4467):
//
//	"Choose up to five {P} worth of modes. You may choose the same mode
//	 more than once.
//	 {P} — Each player sacrifices a creature of their choice.
//	 {P}{P} — Draw a card for each creature that died under your control
//	 this turn.
//	 {P}{P}{P} — Each opponent loses X life, where X is the number of
//	 creature cards in your graveyard."
//
// A pawprint spell (CR 700.2i, #4786). The modes run in printed order
// (CR 608.2c), each occurrence of a repeated one in turn (CR 700.2d), and
// a mode that leaves the edict's prompts open holds the ones after it
// (mode_walk.go). So the creatures sacrificed to the {P} mode have died
// by the time the {P}{P} mode counts them, and are in the graveyard the
// {P}{P}{P} mode counts. Both counts are read as their mode resolves:
// the turn tally's creatures that died under your control (CR 700.4),
// and the creature cards in your graveyard then.
func init() {
	Register(Spec{
		OracleID:     "31b5d0ad-3a4d-4f43-b0a9-c9f7a362680f",
		Name:         "Season of Loss",
		Completeness: CompletenessFull,
		Modes: ChooseUpToPawprints(5, true,
			Pawprint(1, ModeDoing("Each player sacrifices a creature of their choice.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return EachPlayerSacrifices{Match: Creature(), Label: "a creature"}.Apply(ctx)
				})),
			Pawprint(2, ModeDoing("Draw a card for each creature that died under your control this turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{Player: ctx.Controller(), N: ctx.Game.TurnTallyFor(ctx.Controller()).CreaturesDied}.Apply(ctx)
				})),
			Pawprint(3, ModeDoing("Each opponent loses X life, where X is the number of creature cards in your graveyard.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					return eachOpponentLosesLife(ctx.Game, item, g2CardsInYourGraveyard(game.Card.IsCreature)(item, ctx))
				})),
		),
	})
}
