package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coiling Rebirth — Sorcery {2}{B} (Bloomburrow, #4467):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, they draw a card before its other effects.)
//	 Return target creature card from your graveyard to the battlefield.
//	 Then if the gift was promised and that creature isn't legendary,
//	 create a token that's a copy of that creature, except it's 1/1."
//
// The return is the move's continuation, and the copy is made from the
// creature that actually arrived (a card a replacement turned away leaves
// nothing to copy). The 1/1 is a copiable exception on the token (CR 707.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27a91ebd-39f4-432d-b6cc-cd86b4a3979a",
		Name:         "Coiling Rebirth",
		Completeness: CompletenessFull,
		Gift:         GiftACard(),
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 0
			if ctx.GiftPromised() {
				n = 1
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(t.ID, uuid.Nil, false, ctx.TailN(coilingRebirthCopy, n))
			}
			return nil
		},
	})
}
