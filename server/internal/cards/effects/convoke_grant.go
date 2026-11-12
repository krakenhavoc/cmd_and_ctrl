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

// NextSpellHasConvoke is "the next spell you cast this turn has convoke"
// (Wand of the Worldsoul, Flockchaser Phantom; #5647): a one-use
// NextSpellPromise carrying Convoke, given to the controller until end
// of turn. While it stands the caster's next spell is offered the
// convoke tap cost (CR 702.51a); that spell spends it whether or not a
// creature was tapped, and it counts as a spell that has convoke for
// YouCastASpellThatHasConvoke. `from` is the card's name, for the log.
func NextSpellHasConvoke(from string) GrantNextSpellPromise {
	return GrantNextSpellPromise{From: from, Promise: game.NextSpellPromise{
		Convoke: true,
		Text:    "The next spell you cast this turn has convoke.",
	}}
}

// YouCastASpellThatHasConvoke — "whenever you cast a spell that has
// convoke" (Joyful Stormsculptor, Saint Traft and Rem Karolus) and
// "another spell that has convoke" (Kasla, the Broken Halo, whose own
// cast never reaches her: she is not on the battlefield yet). On an
// EventCast: your own cast of a spell that had convoke as it became cast
// (CR 601.2i), whether printed on it, given by a permanent ("Noncreature
// spells you cast have convoke") or given by a next-spell promise. No
// creature need have been tapped for it.
func YouCastASpellThatHasConvoke(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor == source.Controller && g.SpellHadConvokeForEffect(ev.CardID)
}
