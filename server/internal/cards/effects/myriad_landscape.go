package effects

import (
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
// S22 — "two basic land cards that SHARE A LAND TYPE" is now the
// player's call. It used to be resolved by reading the land type off
// the first basic in library order and fetching that type, which is
// neither stronger nor weaker than printed but is not a choice: a
// library whose bottom-most basic was a Plains fetched Plains even
// when the deck wanted Forests.
//
// The clause is the one thing no per-card predicate can express — it
// is a property of the PAIR, not of either card — so it rides on the
// search's Validate hook instead: every basic with a land type is a
// candidate, and the engine rejects a submitted pair that does not
// share one. Taking a single basic, or none, is legal ("up to two").
func init() {
	Register(Spec{
		OracleID:     "2549bc57-9ffb-4053-9f10-f2a5f792b845",
		Name:         "Myriad Landscape",
		Completeness: CompletenessFull,
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
	return SearchLibrary{
		Player:        item.Controller,
		Predicate:     func(c game.Card) bool { return basicLandSubtypeOf(c) != "" },
		Dest:          game.ZoneBattlefield,
		Limit:         2,
		Reveal:        true,
		Shuffle:       true,
		TappedOnEntry: true,
		Reason:        "Myriad Landscape — up to two basic lands that share a land type",
		Validate:      basicsShareALandType,
	}.Apply(NewContext(g, item))
}

// basicsShareALandType is Myriad Landscape's pair constraint. One
// card trivially shares a type with itself, and zero cards is a
// legal "fail to find", so only a two-card pick can fail.
func basicsShareALandType(picked []game.Card) bool {
	if len(picked) < 2 {
		return true
	}
	shared := basicLandSubtypeOf(picked[0])
	if shared == "" {
		return false
	}
	for _, c := range picked[1:] {
		if basicLandSubtypeOf(c) != shared {
			return false
		}
	}
	return true
}

// basicLandSubtypeOf returns the lowercase basic land type of a
// basic land card, or "" when the card is not a basic land or
// carries no basic land type.
//
// Only the five ordinary basic types are recognised. Wastes has no
// basic land type at all (CR 305.6), so a Wastes has nothing to
// share a type with and is correctly not a candidate.
func basicLandSubtypeOf(c game.Card) string {
	if !IsBasicLand(c) {
		return ""
	}
	for _, sub := range []string{"plains", "island", "swamp", "mountain", "forest"} {
		if containsFoldASCII(c.TypeLine, sub) {
			return sub
		}
	}
	return ""
}
