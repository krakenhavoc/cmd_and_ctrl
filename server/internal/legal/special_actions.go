package legal

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// specialActionParams is the special_action wire payload this package
// emits. Strict + auto_tap for the same reason every cast is: the
// enumerator decides affordability here, so the engine never has to
// reject a move a bot was offered (#544).
type specialActionParams struct {
	CardID  string `json:"card_id"`
	Kind    string `json:"kind"`
	Strict  bool   `json:"strict,omitempty"`
	AutoTap bool   `json:"auto_tap,omitempty"`
}

// specialActionMoves enumerates the CR 116.2 special actions the seat
// may take on cards in its own hand — foretell (CR 702.143a) and
// suspend (CR 702.62a).
//
// IT DOES NOT COPY THE SPLIT-SECOND EARLY RETURN that opens
// castMoves and activatedMoves, and that is the whole reason ADR 0062
// Decision 4 wrote this method down before anything implemented it.
// CR 702.61b stops casts and activations of non-mana abilities; a
// special action is neither, so foretelling under a Trickbind is
// legal. Suspend is barred under split second — but by its own
// wording (CR 702.62c imports every restriction on beginning to cast
// the card), which the engine's per-kind timing table says and this
// loop asks rather than assumes.
//
// One question per (card, kind): the engine's own
// SpecialActionTimingOKLocked, so the enumerator and the engine
// cannot drift apart the way two copies of a timing rule always do.
func (e *enumerator) specialActionMoves() {
	g, p := e.g, e.p
	if p == nil || p.Hand == nil {
		return
	}
	for i := range p.Hand.Cards {
		card := p.Hand.Cards[i]
		for _, sa := range game.SpecialActionsFor(game.CatalogKey(card)) {
			if !game.SpecialActionKindBuilt(sa.Kind) {
				continue
			}
			if !g.SpecialActionTimingOKLocked(e.seat, card, sa.Kind) {
				continue
			}
			// The cost is mana and nothing else, and it is the only
			// way a special action can be unaffordable — there is no
			// target to be missing and no choice to be unanswerable.
			// Priced through the same zero spend context the engine
			// pays with, so restricted mana counts here exactly as
			// little as it does there.
			if sa.Cost != "" {
				cost, err := game.ParseCost(sa.Cost)
				if err != nil {
					continue
				}
				if !e.canPay(cost, 0, game.ManaSpendContext{}) {
					continue
				}
			}
			label := sa.Label
			if label == "" {
				label = string(sa.Kind)
			}
			e.add(Move{
				Type:   TypeSpecialAction,
				Player: e.seat,
				Kind:   KindSpecialAction,
				Label:  label + " " + card.Name,
				Source: card.InstanceID,
				Params: mustJSON(specialActionParams{
					CardID:  card.InstanceID.String(),
					Kind:    string(sa.Kind),
					Strict:  true,
					AutoTap: true,
				}),
			})
		}
	}
}
