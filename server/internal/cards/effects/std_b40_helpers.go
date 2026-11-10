package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b40_helpers.go — shared pieces of the Standard-set leftovers, batch
// 40 (Murders at Karlov Manor, #4469).

// targetPlayerExilesFromHand is "target opponent exiles a card from their
// hand" for the resolving item's legal player targets: each picks one card
// of their own hand (a floor of one) and an empty hand exiles nothing
// (CR 609.3). The answer is exileEachChosenCard.
func targetPlayerExilesFromHand(ctx *Context, question string) error {
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
			Question:   question,
			Cards:      cards,
			Min:        1,
			Max:        1,
			ThenKey:    stdB40ExileFromHand,
			LiveZone:   game.ZoneHand,
		})
	}
	return nil
}

// stdB40ExileFromHand is the keyed answer to targetPlayerExilesFromHand.
// An on-disk key: never renamed, never reused.
var stdB40ExileFromHand = game.RegisterChooseCardsThen("choose-cards/std-b40-exile-from-hand", exileEachChosenCard)

// OfAnArtifactYouControl — "activated abilities of artifacts you control":
// the ability's source is an artifact on the battlefield under the
// modifier's controller. An ability that works from another zone, such as
// cycling, is not one (the Forensic Gadgeteer ruling). Clue is an artifact
// type, so a Clue Equipment counts.
func OfAnArtifactYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.FromZone == game.ZoneBattlefield &&
			q.Card.Controller == q.Source.Controller &&
			q.Card.IsArtifact()
	}
}
