package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB122ExileCyclingCardsNotCycled is Abandoned Sarcophagus' "If a card
// that has a cycling ability would be put into your graveyard from anywhere
// and it wasn't cycled, exile it instead": Forbidden Crypt's own-graveyard
// exile for cards (not tokens), narrowed to cards with a cycling ability
// (typecycling counts, CR 702.29e) and not for the one move that IS cycling.
//
// Cycling is a discard paid as the cost of the card's own ability (CR 702.29a:
// "{cost}, Discard this card"), so the move whose discard source is the card
// itself and whose cause is a cost is the cycled one and is left alone. No
// printed card has another self-discard cost beside a cycling ability.
func S82OldB122ExileCyclingCardsNotCycled(label string) game.ReplacementEffect {
	r := GraveyardBecomesExile{YoursOnly: true, CardsOnly: true, Label: label}.Build()
	base := r.AppliesTo
	r.AppliesTo = func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
		if !base(ev, g, src) {
			return false
		}
		moving, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !GapB158HasCyclingAbility(g, ev.CardID, moving) {
			return false
		}
		cycled := ev.Kind == game.RepEventDiscard && ev.DiscardCause == game.DiscardCauseCost && ev.Source == ev.CardID
		return !cycled
	}
	return r
}
