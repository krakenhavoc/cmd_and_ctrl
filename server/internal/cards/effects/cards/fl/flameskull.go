package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flameskull — {1}{R}{R} Creature — Skeleton 3/1 (Adventures in the Forgotten
// Realms):
//
//	"Flying
//	 This creature can't block.
//	 Rejuvenation — When this creature dies, exile it. If you do, exile the
//	 top card of your library. Until the end of your next turn, you may play
//	 one of those cards. (If you cast Flameskull this way, you can't play the
//	 other card, and vice versa.)"
//
// The exile of Flameskull itself is ExileTarget on the source card: it does
// nothing when the card is no longer in the graveyard, or is a new object (CR
// 400.7), and "if you do" is the exile's continuation, run only for a card
// that reached exile. The continuation exiles the top card of the controller's
// library and opens two play permissions (CR 611.2c: the pair is locked), one
// over each card, with the same window (ends with the controller's next turn)
// and ONE budget (BudgetGroup, CastsLeft 1): playing either, a spell or a land
// with the turn's land drop, spends the use of both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f3d935cd-d582-464a-a13b-bb84a53da4a1",
		Name:            "Flameskull",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Flameskull — Rejuvenation: exile it; if you do, exile the top card of your library and play one of those cards until the end of your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ExileTarget{
						Target:   item.SourceCardID,
						ThenTail: ctx.ItemTail(flameskullAfterExile),
					}.Apply(ctx)
				}),
		},
	})
}

// flameskullAfterExile is "if you do": it runs with Flameskull's id as the
// outcome when the card reached exile, and with nothing when it did not. An
// on-disk key: never renamed, never reused.
var flameskullAfterExile = effects.CardTail("flameskull/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	g, item := ctx.Game, ctx.Item
	self := o.IDs[0]
	perm := game.CastPermission{
		Player:      item.Controller,
		BudgetGroup: uuid.New(),
		CastsLeft:   1,
		Duration:    g.UntilEndOfYourNextTurnDuration(item.Controller),
		Source:      item.SourceCardID,
		SourceName:  "Flameskull",
		Label:       "Flameskull — play Flameskull or the exiled card until the end of your next turn",
	}
	// Flameskull is in exile already (this tail runs only once it landed), so its
	// permission is granted at once; then the top card is exiled with the same one.
	if err := g.ExileCardWithPermissionForEffect(self, perm); err != nil {
		return err
	}
	_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, perm)
	return err
})
