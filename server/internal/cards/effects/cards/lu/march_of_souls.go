package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// March of Souls — Sorcery {4}{W} (Planeshift, S82 OLD b644, #5615):
//
//	"Destroy all creatures. They can't be regenerated. For each creature destroyed
//	 this way, its controller creates a 1/1 white Spirit creature token with flying."
//
// Ceaseless Conflict's batched sweep: every creature is destroyed as one event (CR
// 700.4), ignoring regeneration shields (CR 701.19c), and the tail is handed the
// pre-move copies of the creatures that really were destroyed, so an indestructible
// one pays nothing. Each such creature's controller, as it last existed on the
// battlefield (CR 608.2h), makes one token; the tokens are made per controller in
// turn order from the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bca7101b-e74b-42af-ac63-1b3defb4137d",
		Name:         "March of Souls",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Creature(), CantBeRegenerated: true, ThenTail: ctx.ItemTail(marchOfSoulsSpirits)}.Apply(ctx)
		},
	})
}

// marchOfSoulsSpirits makes a 1/1 white Spirit with flying for each creature destroyed
// this way, under its controller's control. An on-disk key.
var marchOfSoulsSpirits = effects.MassTail("mass/march-of-souls-spirits", func(ctx *effects.Context, swept []game.Card, _ int) error {
	for _, seat := range effects.SeatsFromActive(ctx.Game) {
		n := 0
		for _, c := range swept {
			if c.Controller == seat {
				n++
			}
		}
		if n == 0 {
			continue
		}
		if err := (effects.CreateToken{Controller: seat, Template: effects.TokenCard("1/1 white Spirit with flying"), N: n}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
