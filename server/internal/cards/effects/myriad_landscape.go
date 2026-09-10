package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myriad Landscape — Land (EDHREC rank 31):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Search your library for up to two
//	 basic land cards that share a land type, put them onto the
//	 battlefield tapped, then shuffle."
//
// A colourless land that turns into two lands. It is played in
// essentially every deck that can afford a tapped land, which is why
// it outranks most of the duals.
//
// Three components on the sacrifice cost — mana, tap and the land
// itself — the shape Mind Stone established.
//
// SANDBOX SIMPLIFICATION — "two basic land cards THAT SHARE A LAND
// TYPE" is resolved by looking at the first basic in library order,
// taking its land type, and fetching up to two basics of that type.
// A real chooser would let the controller name the type; the
// deterministic first-match rule is the same one every SearchLibrary
// card in the catalog uses, and this card is the one where it bites
// hardest — a library whose bottom-most basic is a Plains fetches
// Plains even when the deck wanted Forests. That is neither stronger
// nor weaker than printed, just less controllable, and a search
// chooser fixes it for every card at once.
func init() {
	Register(Spec{
		OracleID:     "2549bc57-9ffb-4053-9f10-f2a5f792b845",
		Name:         "Myriad Landscape",
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:  "{2}, {T}, Sacrifice this land: Search your library for up to two basic land cards that share a land type, put them onto the battlefield tapped, then shuffle.",
			Cost:   Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: myriadLandscapeFetch,
		}},
	})
}

func myriadLandscapeFetch(g *game.Game, item *game.StackItem) error {
	subtype := firstBasicLandSubtype(g, item.Controller)
	if subtype == "" {
		// No basic in the library at all. The cost is still paid
		// (CR 601.2h) — the search simply finds nothing.
		return nil
	}
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     func(c game.Card) bool { return IsBasicLand(c) && containsFoldASCII(c.TypeLine, subtype) },
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
	}.Apply(NewContext(g, item))
}

// firstBasicLandSubtype returns the lowercase basic land type of the
// first basic in the player's library, walking the slice in the same
// order SearchLibrary does. Empty string when the library holds no
// basic land.
//
// Only the five ordinary basic types are recognised. Wastes has no
// basic land type at all (CR 305.6), so a Wastes-only library
// correctly finds nothing to share a type with.
func firstBasicLandSubtype(g *game.Game, playerID uuid.UUID) string {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Library == nil {
		return ""
	}
	for _, c := range p.Library.Cards {
		if !IsBasicLand(c) {
			continue
		}
		for _, sub := range []string{"plains", "island", "swamp", "mountain", "forest"} {
			if containsFoldASCII(c.TypeLine, sub) {
				return sub
			}
		}
	}
	return ""
}
