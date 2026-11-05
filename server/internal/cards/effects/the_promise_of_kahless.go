package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Promise of Kahless — Sorcery {4}{B} (Star Trek Commander, #4974):
//
//	"You may sacrifice a Klingon or Warrior. If you do, return all other
//	 Klingon and Warrior creature cards from your graveyard to the
//	 battlefield."
//
// The sacrifice is chosen as the spell resolves (maySacrificeOne, CR
// 608.2c), over the permanents its controller controls then with the
// subtype Klingon or Warrior (post-layer, so a changeling counts), with a
// floor of zero for the "may". Declining, or controlling none, does
// nothing. "If you do" follows only a permanent that really left (a
// commander sent to the command zone still counts, CR 701.21a).
//
// The return is read after the sacrifice has settled: every creature card
// in the caster's graveyard that is a Klingon or a Warrior, except the
// sacrificed card itself ("other"), entering together under the caster's
// control (CR 603.6a). The words are read through the spell's text (CR
// 612.2); the spell is still on the stack while the sacrifice is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5866d5e-ee03-4bca-9964-0d391eae9011",
		Name:         "The Promise of Kahless",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			klingon, warrior := ctx.TextSubtype("Klingon"), ctx.TextSubtype("Warrior")
			return maySacrificeOne(ctx.Game, item, "The Promise of Kahless — sacrifice a Klingon or Warrior?",
				Or(HasSubtype(klingon), HasSubtype(warrior)), false, promiseOfKahlessReturn)
		},
	})
}

// promiseOfKahlessReturn is "return all other Klingon and Warrior creature
// cards from your graveyard to the battlefield", once the sacrifice has
// happened. An on-disk key: never renamed, never reused.
var promiseOfKahlessReturn = RegisterSacrificedOne("prompt-run/the-promise-of-kahless-sacrifice-return", func(ctx *Context, sacrificed uuid.UUID) error {
	klingon, warrior := ctx.TextSubtype("Klingon"), ctx.TextSubtype("Warrior")
	ids := graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
		return c.InstanceID != sacrificed && c.IsCreature() && (c.HasSubtypeWord(klingon) || c.HasSubtypeWord(warrior))
	})
	if len(ids) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
})
