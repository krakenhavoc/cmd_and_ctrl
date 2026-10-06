package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Binding Negotiation — Sorcery {1}{B}:
//
//	"Target opponent reveals their hand. You may choose a nonland card
//	 from it. If you do, they discard it. Otherwise, you may put a
//	 face-up exiled card they own into their graveyard."
//
// The optional revealed-hand pick (#2115, ADR 0116's 2026-10-05
// amendment): the whole table sees the hand (CR 701.20a), and you may
// choose a nonland card or nothing. A chosen card is discarded.
// Otherwise — you chose nothing, or there was nothing to choose — you
// may choose one face-up card in exile that the player owns, and it is
// put into its owner's graveyard. A face-down exiled card is not
// offered (CR 406.3); its identity is hidden even from you.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86dc3a41-4fb9-4b1f-8279-2b5edbc7e170",
		Name:         "Binding Negotiation",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChooseFromRevealedHand{
				Player:   TargetedPlayer(ctx),
				Filter:   Nonland(),
				Label:    "nonland card",
				Optional: true,
				Then:     bindingNegotiationOtherwise,
			}.Apply(ctx)
		},
	})
}

// bindingNegotiationOtherwise is "Otherwise, you may put a face-up
// exiled card they own into their graveyard."
var bindingNegotiationOtherwise = RevealedPickThen("revealed-pick/binding-negotiation-exiled-to-graveyard",
	func(ctx *Context, pick game.RevealedPick) error {
		if len(pick.Chosen) > 0 {
			return nil
		}
		g := ctx.Game
		var offer []uuid.UUID
		if g.Exile != nil {
			for _, c := range g.Exile.Cards {
				if c.Owner == pick.FromPlayer && !c.FaceDown {
					offer = append(offer, c.InstanceID)
				}
			}
		}
		if len(offer) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    ctx.Controller(),
			FromPlayer: pick.FromPlayer,
			Source:     pick.Source,
			Question:   "Binding Negotiation — you may put a face-up exiled card they own into their graveyard",
			Cards:      offer,
			Min:        0,
			Max:        1,
			Zone:       game.ZoneExile,
			Then: func(g *game.Game, picked []uuid.UUID) error {
				return g.PutCardsIntoGraveyardThenForEffect(picked, nil)
			},
		})
		return nil
	})
