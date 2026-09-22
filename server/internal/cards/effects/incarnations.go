package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// incarnations.go — #1221: the Judgment incarnation cycle, the family
// that wants CR 113.6c.
//
//	Anger   {3}{R} 2/2  Haste.
//	                    "As long as this card is in your graveyard and
//	                     you control a Mountain, creatures you control
//	                     have haste."
//	Wonder  {3}{U} 2/2  Flying / an Island / flying.
//	Brawn   {3}{G} 3/3  Trample / a Forest / trample.
//	Valor   {3}{W} 2/2  First strike / a Plains / first strike.
//
// Four cards, one shape, and the shape is the seam: a static ability
// whose SOURCE is a card in a graveyard. The printed keyword on the
// body is an ordinary keyword the importer already carries; the
// anthem is the part that had nowhere to live until
// StaticAbility.Zones and the layer pass's declared-zone gather
// (game/static_zones.go).
//
// Two details the cards make you get right, and both fall out of
// CR 108.4 rather than being decided here:
//
//   - "you" is the card's OWNER. A card in a graveyard has no
//     controller, so an Anger milled out of an opponent's library
//     gives THAT opponent's creatures haste and not yours. The
//     gather sets Controller = Owner on the source it binds, which
//     is what makes `source.Controller` below mean the right player.
//   - The anthem does NOT apply while the incarnation is on the
//     BATTLEFIELD. The declared zone list is the list, so a
//     battlefield Anger is a 2/2 with haste and nothing else — which
//     is exactly what the card says and the reason the printed
//     keyword and the graveyard clause are two separate lines of
//     rules text.

// incarnationAnthem builds one incarnation's graveyard static:
// "As long as this card is in your graveyard and you control a
// <land>, creatures you control have <keyword>."
//
// `keyword` is a canonical keyword token ("haste", "flying",
// "trample", "first strike"); `land` is the basic land TYPE the
// clause names ("Mountain", "Island", "Forest", "Plains").
//
// Layer 6, because granting an ability is what layer 6 is, and
// therefore commutative with every other grant — which is why the
// gather's approximate timestamp for a graveyard source is
// unobservable for this whole family.
func incarnationAnthem(keyword, land string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer6Ability,
		// CR 113.6c: this ability functions from a GRAVEYARD and
		// nowhere else. #1221.
		Zones: []game.ZoneKind{game.ZoneGraveyard},
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if !target.IsCreature() || target.Controller != source.Controller {
				return false
			}
			return controlsLandType(g, source.Controller, land)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if !eotHasAbility(c.Abilities, keyword) {
				c.Abilities = append(c.Abilities, keyword)
			}
		},
	}
}

// controlsLandType is the clause's other half: "and you control a
// Mountain". A basic land TYPE, not the basic supertype — a Sacred
// Foundry is a Mountain and turns Anger on, and so is an animated
// land or one a type-changing effect has retyped.
//
// Walks g.Battlefield directly rather than through
// BattlefieldCardsForEffect, which copies the whole pile: this runs
// inside the layer pass's AppliesTo, once per candidate creature per
// recompute, and a copy per call would make the pass quadratic in the
// board.
//
// HasSubtype is a PASSIVE read — it uses the cached resolution when
// one is present and the printed type line otherwise — which is the
// invariant an AppliesTo predicate has to keep (see card.go on
// Effective()).
func controlsLandType(g *game.Game, controller uuid.UUID, land string) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.HasSubtype(land) {
			return true
		}
	}
	return false
}
