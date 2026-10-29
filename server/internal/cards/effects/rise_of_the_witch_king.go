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
// EachPlayerSacrificesThenKeyedForEffect with the controller included —
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
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.EachPlayerSacrificesThenKeyedForEffect(
				ctx.Source(), uuid.Nil,
				sacrificeSpec("a creature", Creature()),
				"Rise of the Witch-king — sacrifice a creature",
				riseOfTheWitchKingReturn, ctx.RunData([]uuid.UUID{ctx.Controller()}))
		},
	})
}

// riseOfTheWitchKingReturn is "if you sacrificed a creature this way, you
// may return another permanent card" once the sacrifices have landed, keyed
// (#4290): the controller is the run's carry, Rise itself its source.
var riseOfTheWitchKingReturn = RunThen("prompt-run/rise-of-the-witch-king-return", func(ctx *Context, r game.RunDone) error {
	if len(r.Carry) == 0 || !r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	return ReturnChosenFromGraveyard{
		Player:   r.Carry[0],
		Question: "Rise of the Witch-king — you may return another permanent card from your graveyard to the battlefield",
		Except:   r.Source,
		Max:      1,
	}.Apply(ctx)
})
