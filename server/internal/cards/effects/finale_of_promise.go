package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Finale of Promise — Sorcery {X}{R}{R}:
//
//	"You may cast up to one target instant card and/or up to one target
//	 sorcery card from your graveyard each with mana value X or less
//	 without paying their mana costs. If a spell cast this way would be
//	 put into your graveyard, exile it instead. If X is 10 or more, copy
//	 each of those spells twice. You may choose new targets for the
//	 copies."
//
// Two target clauses (CR 601.2c), each up to one and bound to the announced
// X with WithManaValueAtMostX (Here Comes a New Hero!'s binding, re-checked
// on resolution, CR 608.2b). Each pick gets its own free, exile-instead
// permission, so casting one leaves the other open, and the casts are offered
// during the resolution (CR 608.2g).
//
// "If X is 10 or more, copy each of those spells twice" is a follow-up on
// each cast, granted only when X is 10 or more: it copies the spell that is
// now on the stack twice, each copy offering new targets (CR 707.10c). A
// card that is not cast is not a spell, so it is not copied. The copies
// are made as each spell is cast, before the next one is, which is the
// order the casts and copies could be answered in no differently: nothing
// resolves in between.
func init() {
	Register(Spec{
		OracleID:     "d6fafb50-9531-4fbf-bb1e-ebb4dd39281c",
		Name:         "Finale of Promise",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("up to one target instant card with mana value X or less from your graveyard",
				Instant(), YouOwn()).WithCount(0, 1).WithManaValueAtMostX(),
			TargetCardInGraveyard("up to one target sorcery card with mana value X or less from your graveyard",
				Sorcery(), YouOwn()).WithCount(0, 1).WithManaValueAtMostX(),
		),
		OnResolve: finaleOfPromiseResolve,
	})
}

// finaleOfPromiseCopyTwice copies a spell cast this way twice, each with a
// chance at new targets. A spell that left the stack in between is not
// copied (it cannot have: nothing resolves between the cast and this).
var finaleOfPromiseCopyTwice = game.RegisterCastFollowUp("finale-of-promise/copy-twice",
	func(g *game.Game, f game.CastFollowUp) error {
		for i := 0; i < 2; i++ {
			if err := g.CopySpellForEffect(f.Spell, f.Player, true, nil); err != nil && err != game.ErrCardNotFound {
				return err
			}
		}
		return nil
	})

func finaleOfPromiseResolve(item *game.StackItem, ctx *Context) error {
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	var followUp game.CastFollowUpKey
	if ctx.X() >= 10 {
		followUp = finaleOfPromiseCopyTwice
	}
	cmdr2S42FreeGraveyardCastsThen(ctx, "Finale of Promise", freeInstantOrSorcery(), true, false, followUp, ids)
	return nil
}
