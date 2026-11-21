package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neutralize the Guards — Instant {2}{B} (Outlaws of Thunder Junction,
// #4468):
//
//	"Creatures target opponent controls get -1/-1 until end of turn.
//	 Surveil 2. (Look at the top two cards of your library, then put any
//	 number of them into your graveyard and the rest on top of your
//	 library in any order.)"
//
// Gnashing of Teeth's second bullet held to an opponent, then the
// Surveil primitive. The creatures are those the opponent controls as
// the spell resolves (CR 611.2c), so one that comes under their control
// later is not shrunk. An opponent who has left the game by then makes
// the spell's only target illegal, so it does nothing, the surveil
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b793141d-581c-4cc7-b09f-1fe8bd5536c5",
		Name:         "Neutralize the Guards",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.ControlledByPlayer(t.ID)), Power: -1, Toughness: -1,
				Label: "Neutralize the Guards — -1/-1",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
