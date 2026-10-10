package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rise of the Witch-king — Sorcery {2}{B}{G} (EDHREC rank 1924):
//
//	"Each player sacrifices a creature of their choice. If you
//	 sacrificed a creature this way, you may return another permanent
//	 card from your graveyard to the battlefield."
//
// An edict that pays you back. The sacrifice is
// EachPlayerSacrificesThenForEffect with the controller included —
// every player picks their own, one prompt each, APNAP, and a player
// with no creature is skipped (CR 701.21a).
//
// "IF YOU SACRIFICED A CREATURE THIS WAY" is the run's answer for the
// controller's seat (#1019, ADR 0013 §5x), read once every asked seat
// has answered AND the permanents they named have finished moving — so
// a sacrificed commander's CR 903.9 prompt holds the payout too, and a
// commander that takes the command zone still counts, because CR
// 701.17a's sacrifice is the move off the battlefield.
//
// "YOU MAY RETURN ANOTHER PERMANENT CARD" is chosen then, on
// resolution, from the graveyard as it is after the sacrifices (#2863,
// ADR 0013 amendment of 2026-10-09): ReturnChosenFromGraveyard, with a
// floor of zero. Nothing is targeted, so opponents see no pick at cast
// and the spell cannot be countered for want of one. "Another" is
// another than Rise itself, so the creature you just sacrificed is a
// legal choice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c86541c-3601-4a38-8872-39705e41303a",
		Name:         "Rise of the Witch-king",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller, self := ctx.Controller(), ctx.Source()
			return ctx.Game.EachPlayerSacrificesThenForEffect(
				self, uuid.Nil,
				sacrificeSpec("a creature", Creature()),
				"Rise of the Witch-king — sacrifice a creature",
				func(g *game.Game, sacrificed game.PromptedSacrifices) error {
					if !sacrificed.Sacrificed(controller) {
						return nil
					}
					// A fresh Context bound to the same stack item: an
					// undo restores the game's fields in place, so the
					// *Game captured when the prompts went up can be
					// the wrong object by the time they are answered
					// (resumeClause's contract).
					return ReturnChosenFromGraveyard{
						Player:   controller,
						Question: "Rise of the Witch-king — you may return another permanent card from your graveyard to the battlefield",
						Except:   self,
						Max:      1,
					}.Apply(NewContext(g, item))
				})
		},
	})
}
