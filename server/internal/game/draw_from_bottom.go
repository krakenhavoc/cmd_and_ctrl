package game

import "github.com/google/uuid"

// draw_from_bottom.go — "You draw cards from the bottom of your library
// rather than the top" (River Song's Meet in Reverse, #6089).
//
// Not a replacement: the draw still happens, as a draw (CR 121.1 says a
// player draws "the top card of their library" by default, and this static
// changes which card that is for its controller). So every draw
// replacement still sees it, EventDrawCard still fires, and the only change
// is the one card actuallyDrawCardLocked takes out of the library: the
// bottom one, the first element of Library.Cards (Zone.Bottom).
//
// Judged when each individual draw is performed, through
// catalogAbilityKeyOf, so a permanent that lost its abilities, or one no
// longer under the drawer's control, changes nothing.

// CatalogDrawsFromBottom reports whether a battlefield permanent with this
// catalog key makes its controller draw from the bottom of their library.
// Populated from effects.Spec.DrawsFromBottom. A nil hook finds none.
var CatalogDrawsFromBottom func(key string) bool

// drawsFromBottomLocked reports whether player draws from the bottom of
// their library right now.
//
// Caller must hold g.mu.
func (g *Game) drawsFromBottomLocked(player uuid.UUID) bool {
	if g.Battlefield == nil || CatalogDrawsFromBottom == nil || player == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogDrawsFromBottom(key) {
			return true
		}
	}
	return false
}

// popDrawnCardLocked takes the card p's next draw takes out of their
// library: the top one, or the bottom one under DrawsFromBottom.
//
// Caller must hold g.mu.
func (g *Game) popDrawnCardLocked(p *Player) (Card, error) {
	if !g.drawsFromBottomLocked(p.ID) {
		return p.Library.PopTop()
	}
	bottom, err := p.Library.Bottom()
	if err != nil {
		return Card{}, err
	}
	return p.Library.Remove(bottom.InstanceID)
}
