package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2B21HasSubtype reports whether the card has the subtype, read through
// the layered view.
func cmdr2B21HasSubtype(c game.Card, sub string) bool {
	for _, s := range c.Effective().Subtypes {
		if strings.EqualFold(s, sub) {
			return true
		}
	}
	return false
}

// cmdr2B21HandAtMostOneDiscount is "costs {2} less to activate if you have one
// or fewer cards in hand": 2 when the activating player's hand holds at most
// one card, else 0.
func cmdr2B21HandAtMostOneDiscount(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || (p.Hand != nil && len(p.Hand.Cards) > 1) {
		return 0
	}
	return 2
}

// cmdr2B21EquippedCreature is "equipped creature" as an ability resolves: the
// creature on the battlefield the Equipment `source` is attached to now. False
// when the Equipment is gone, unattached, or its host is no longer a creature
// on the battlefield.
func cmdr2B21EquippedCreature(g *game.Game, source uuid.UUID) (game.Card, bool) {
	if !onBattlefield(g, source) {
		return game.Card{}, false
	}
	eq, ok := g.LookupCardForEffect(source)
	if !ok || eq.AttachedTo.Kind != game.TargetCard {
		return game.Card{}, false
	}
	if !onBattlefield(g, eq.AttachedTo.ID) {
		return game.Card{}, false
	}
	host, ok := g.LookupCardForEffect(eq.AttachedTo.ID)
	if !ok || !host.IsCreature() {
		return game.Card{}, false
	}
	return host, true
}

// cmdr2B21BasePTCount counts the creatures `controller` controls whose base
// power and toughness satisfy `match` (CR 208.4b: the values after layer 7b).
// It walks the battlefield directly, so it is safe inside a layer recompute.
func cmdr2B21BasePTCount(g *game.Game, controller uuid.UUID, match func(power, toughness int) bool) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		if match(c.BasePower(), c.BaseToughness()) {
			n++
		}
	}
	return n
}
