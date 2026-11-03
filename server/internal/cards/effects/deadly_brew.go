package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadly Brew — Sorcery {B}{G}:
//
//	"Each player sacrifices a creature or planeswalker of their choice.
//	 If you sacrificed a permanent this way, you may return another
//	 permanent card from your graveyard to your hand."
//
// Rise of the Witch-king's edict-then-payout: every player, the caster
// included, picks their own creature or planeswalker in APNAP order, a
// player with none is skipped, and the payout reads the run's answer for
// the caster's seat once every sacrifice has landed (so a commander that
// took the command zone still counts as sacrificed).
//
// "Another permanent card" is another than the permanent you sacrificed:
// Deadly Brew is a sorcery, not a permanent card, so the only other
// object the clause can mean is that one. The card it sacrificed is left
// out of the choice; any other permanent card in your graveyard may be
// returned, or none. Nothing is targeted, so the pick is made on
// resolution from the graveyard as it is after the sacrifices.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b831d3f8-07ce-4172-ba29-213cac414c9a",
		Name:         "Deadly Brew",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.EachPlayerSacrificesThenKeyedForEffect(
				ctx.Source(), uuid.Nil,
				sacrificeSpec("a creature or planeswalker", Or(Creature(), Planeswalker())),
				"Deadly Brew — sacrifice a creature or planeswalker",
				deadlyBrewReturn, ctx.RunData([]uuid.UUID{ctx.Controller()}))
		},
	})
}

// deadlyBrewReturn is "if you sacrificed a permanent this way, you may
// return another permanent card from your graveyard to your hand", keyed:
// the caster is the run's carry.
var deadlyBrewReturn = RunThen("prompt-run/deadly-brew-return", func(ctx *Context, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	you := r.Carry[0]
	sacrificed := r.Sacrifices().By(you)
	if len(sacrificed) == 0 {
		return nil
	}
	return ReturnChosenFromGraveyard{
		Player:   you,
		Question: "Deadly Brew — you may return another permanent card from your graveyard to your hand",
		Match:    func(c game.Card) bool { return c.IsPermanent() },
		Except:   sacrificed[0],
		Max:      1,
		ToHand:   true,
	}.Apply(ctx)
})
