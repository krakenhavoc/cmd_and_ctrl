package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB286OpponentExilesTwoFromHand is "Target opponent exiles two cards from
// their hand" for an activated ability (Vessel of Malignity): the opponent
// picks their own two cards (all of them with fewer than two, none from an
// empty hand) and each is exiled through Ruthless Negotiation's keyed answer.
func OldB286OpponentExilesTwoFromHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
			return nil
		}
		cards := make([]uuid.UUID, 0, p.Hand.Size())
		for _, c := range p.Hand.Cards {
			cards = append(cards, c.InstanceID)
		}
		n := min(2, len(cards))
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:    t.ID,
			FromPlayer: t.ID,
			Source:     item.SourceCardID,
			Question:   "Vessel of Malignity — exile two cards from your hand",
			Cards:      cards,
			Min:        n,
			Max:        n,
			ThenKey:    RuthlessNegotiationExiled,
			LiveZone:   game.ZoneHand,
		})
		return nil
	}
	return nil
}
