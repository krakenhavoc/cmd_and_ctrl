package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Library of Lat-Nam — Sorcery {4}{U}:
//
//	"An opponent chooses one —
//	 • You draw three cards at the beginning of the next turn's upkeep.
//	 • You search your library for a card, put that card into your hand,
//	   then shuffle."
//
// The opponent chooses the bullet as the spell is cast (CR 700.2e,
// OpponentChoosesOne), and the caster picks which opponent when there is
// more than one. The first bullet is a CR 603.7 delayed trigger at the
// very next upkeep at the table, whoever's turn it is ("the next turn's",
// not "your next"), controlled by the caster (CR 603.7d). The second is
// Demonic Tutor's search.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67f68a0f-bd3d-47ba-ad99-95c1a2acd170",
		Name:         "Library of Lat-Nam",
		Completeness: effects.CompletenessFull,
		Modes: effects.OpponentChoosesOne(
			effects.ModeWithPurpose(effects.Mode("You draw three cards at the beginning of the next turn's upkeep."), game.Purpose{Draws: 3}),
			effects.ModeWithPurpose(effects.Mode("You search your library for a card, put that card into your hand, then shuffle."), game.Purpose{Tutors: 1}),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.ScheduleDelayedTrigger{
					At:    game.StepUpkeep,
					Label: "Library of Lat-Nam — draw three cards",
					Body:  libraryOfLatNamDrawBody,
				}.Apply(ctx)
			}
			if !ctx.HasMode(1) {
				return nil
			}
			return effects.SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Library of Lat-Nam — search your library for a card",
			}.Apply(ctx)
		},
	})
}

// libraryOfLatNamDrawBody is the delayed trigger's effect: its controller,
// the caster, draws three cards. An on-disk key: never renamed, never
// reused.
var libraryOfLatNamDrawBody = game.SimpleDelayedBody("library-of-lat-nam/draw-three", func(g *game.Game, item *game.StackItem) error {
	return effects.DrawCards{Player: item.Controller, N: 3}.Apply(effects.NewContext(g, item))
})
