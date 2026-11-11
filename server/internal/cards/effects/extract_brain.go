package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Extract Brain — Sorcery {X}{U}{B} (Outlaws of Thunder Junction
// Commander, #5457):
//
//	"Target opponent chooses X cards from their hand. Look at those cards.
//	 You may cast a spell from among them without paying its mana cost."
//
// The opponent chooses, from their own hand and seen by nobody else: X
// cards, or the whole hand when it holds X or fewer (then there is
// nothing to choose). The caster alone looks at the chosen cards
// (game.LookAtHandCardsForEffect); it is not a reveal. The cast is part
// of the resolution (CR 608.2g): the resolution_cast prompt over the
// chosen cards that have a spell to cast, out of the opponent's hand
// (#5457, game.foreignHandForCastLocked), free, at any timing, one spell
// only. Additional costs are still paid and an X in the stolen spell's
// cost is 0 (CR 107.3b). A card that is not cast stays in its owner's
// hand, and the caster has still seen it. A land is never a spell.
//
// X of zero, or an empty hand, does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bb0eb9c2-c26d-452d-9e02-fc0bec2655c4",
		Name:         "Extract Brain",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				extractBrainChoose(ctx, t.ID, ctx.X())
				break
			}
			return nil
		},
	})
}

// extractBrainChoose asks the opponent for X cards from their hand, or
// takes the whole hand when it has no more than X.
func extractBrainChoose(ctx *Context, opponent uuid.UUID, x int) {
	g := ctx.Game
	hand := allHandCardIDs(g, opponent)
	if x <= 0 || len(hand) == 0 {
		return
	}
	if len(hand) <= x {
		extractBrainLookAndOffer(g, ctx.Controller(), opponent, ctx.Source(), hand)
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    opponent,
		FromPlayer: opponent,
		Source:     ctx.Source(),
		Question:   "Extract Brain — choose the cards from your hand your opponent looks at",
		Cards:      hand,
		Min:        x,
		Max:        x,
		ThenKey:    extractBrainChosen,
		Carry:      []uuid.UUID{ctx.Controller()},
	})
}

// extractBrainChosen is the opponent's choice answered: Carry is the
// spell's controller. An on-disk key: never renamed, never reused.
var extractBrainChosen = game.RegisterChooseCardsThen("choose-cards/extract-brain-chosen", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) > 0 {
		extractBrainLookAndOffer(g, r.Carry[0], r.Chooser, r.Source, r.Picked)
	}
	return nil
})

// extractBrainLookAndOffer is "Look at those cards. You may cast a spell
// from among them without paying its mana cost."
func extractBrainLookAndOffer(g *game.Game, controller, opponent, source uuid.UUID, chosen []uuid.UUID) {
	looked := g.LookAtHandCardsForEffect(controller, opponent, chosen)
	perm := g.FreeCastFromHandPermission(controller, source, "Extract Brain")
	candidates := freeCastCandidates(g, &perm, looked)
	if len(candidates) == 0 {
		return
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       source,
		Reason:       "Extract Brain",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: declineLeaveInTheirHand,
	})
}

// declineLeaveInTheirHand is the decline of a free cast out of another
// player's hand; the client words what not casting does from it
// (mayCast.ts whereItStays).
const declineLeaveInTheirHand = "Leave it in their hand"
