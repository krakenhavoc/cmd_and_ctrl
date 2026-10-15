package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// convoke_grant.go — a permanent's "<spells> you cast have convoke"
// (#2137, #3245; game/granted_tap_cost.go). Append-only.
//
//	GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveConvoke(
//	    "Spells you cast from exile have convoke.", FromZone(game.ZoneExile))},   // Hoarding Broodlord
//	GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveConvoke(
//	    "Creature spells you cast have convoke.", SpellIs(Creature()))},          // Dazzling Theater
//
// In a Room door, put it in Door.GrantedTapCosts and the door gates it.

// ConvokeGrantFilter is which spells a convoke grant reaches: the card
// being cast and the zone it is cast from.
type ConvokeGrantFilter = func(g *game.Game, caster uuid.UUID, spell game.Card, from game.ZoneKind) bool

// SpellsYouCastHaveConvoke is the grant. A nil filter is every spell.
func SpellsYouCastHaveConvoke(label string, spells ConvokeGrantFilter) game.GrantedTapCost {
	return game.GrantedTapCost{Spells: spells, Cost: Convoke(), Label: label}
}

// FromZone is "spells you cast from <zone>".
func FromZone(zone game.ZoneKind) ConvokeGrantFilter {
	return func(_ *game.Game, _ uuid.UUID, _ game.Card, from game.ZoneKind) bool { return from == zone }
}

// SpellIs is "<kind of> spells you cast", read against the card.
func SpellIs(pred CardPredicate) ConvokeGrantFilter {
	return func(g *game.Game, caster uuid.UUID, spell game.Card, _ game.ZoneKind) bool {
		return pred(g, caster, spell)
	}
}

// SpellsYouCastHaveImprovise is the grant for "<spells> you cast have
// improvise" (Ironheart, Clever Champion; CR 702.126): convoke's twin
// over artifacts. A nil filter is every spell.
func SpellsYouCastHaveImprovise(label string, spells ConvokeGrantFilter) game.GrantedTapCost {
	return game.GrantedTapCost{Spells: spells, Cost: Improvise(), Label: label}
}
