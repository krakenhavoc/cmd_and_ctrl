package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ruthless Negotiation — Sorcery {B} (Bloomburrow, #4467):
//
//	"Target opponent exiles a card from their hand. If this spell was
//	 cast from a graveyard, draw a card.
//	 Flashback {4}{B}"
//
// The target opponent picks the card (Karn Liberated's pick from their
// own hand, a floor of one, since "exiles a card" is not a "may"). An
// empty hand has nothing to exile (CR 609.3). The draw is the second
// step and waits for that pick (CR 608.2c, #4155). "Cast from a
// graveyard" reads CastFromZone, as Increasing Vengeance and Moment of
// Glory do, not the cost that was paid. Flashback is the shared seam,
// which exiles the card as it leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a82ea8f2-e576-46cd-aa08-9c98d6029db4",
		Name:             "Ruthless Negotiation",
		Completeness:     CompletenessFull,
		Targets:          TargetPlayer("target opponent", Opponent()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{4}{B}")},
		OnResolve: ResolveSteps(
			Instruction(ruthlessNegotiationExile),
			Instruction(func(ctx *Context) error {
				if ctx.Item.CastFromZone != game.ZoneGraveyard {
					return nil
				}
				return DrawCards{N: 1}.Apply(ctx)
			}),
		),
	})
}

// ruthlessNegotiationExile asks the target opponent which card of their
// hand to exile.
func ruthlessNegotiationExile(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := ctx.Game.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			continue
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    t.ID,
			FromPlayer: t.ID,
			Source:     ctx.Item.SourceCardID,
			Question:   "Ruthless Negotiation — exile a card from your hand",
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    ruthlessNegotiationExiled,
			LiveZone:   game.ZoneHand,
		})
	}
	return nil
}

// ruthlessNegotiationExiled is the keyed answer (#4253). An on-disk key:
// never renamed, never reused.
var ruthlessNegotiationExiled = game.RegisterChooseCardsThen("choose-cards/ruthless-negotiation-exile", exileEachChosenCard)
