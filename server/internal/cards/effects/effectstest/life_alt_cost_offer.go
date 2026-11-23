package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// life_alt_cost_offer_test.go — #695, the card half. Force of Will
// and Snuff Out are the two printed offers with a life component, and
// the bug was that the view offered them to a player who could not
// pay: the picker listed "Pay 4 life", the player chose it, and
// CastSpell answered ErrInvalidParam.
//
// So each card is asserted twice on the same board — the offer the
// VIEW shows, and what CastSpell does with it — because the whole
// point is that the two agree. The engine-level boundary cases live
// in internal/game/alternative_cost_payable_test.go and the projection
// in internal/protocol; these hold the claim to the real cards.

// visibleHandCard is handCardFull plus the visibility the projection
// needs: a card its owner cannot SEE is redacted, and redaction
// clears the alternative-cost offers along with the mana cost — so a
// fixture without KnownBy would "pass" the hidden-offer assertions
// for the wrong reason.
func VisibleHandCard(p *game.Player, name, typeLine, manaCost, oracle string, colors []string) uuid.UUID {
	id := HandCardFull(p, name, typeLine, manaCost, oracle, colors)
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].KnownBy = map[uuid.UUID]bool{p.ID: true}
		}
	}
	return id
}

// offeredKeys reads the alternative-cost keys the view stamps on a
// card in the seat's own hand.
func OfferedKeys(t *testing.T, g *game.Game, seat *game.Player, id uuid.UUID) []string {
	t.Helper()
	v := protocol.ViewOfGameFor(g, seat.ID.String())
	for _, s := range v.Seats {
		if s.ID != seat.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID != id.String() {
				continue
			}
			keys := make([]string, 0, len(c.AlternativeCosts))
			for _, offer := range c.AlternativeCosts {
				keys = append(keys, offer.Key)
			}
			return keys
		}
	}
	t.Fatalf("the hand card is missing from its owner's view")
	return nil
}

func HasKey(keys []string, want string) bool {
	for _, k := range keys {
		if k == want {
			return true
		}
	}
	return false
}
