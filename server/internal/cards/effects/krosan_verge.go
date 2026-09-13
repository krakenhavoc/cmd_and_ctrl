package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Krosan Verge — Land (EDHREC rank 973):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {2}, {T}, Sacrifice this land: Search your library for a Forest
//	 card and a Plains card, put them onto the battlefield tapped,
//	 then shuffle."
//
// The Selesnya ramp land: two lands for three mana, and "a Forest
// card" is any land with the subtype — a Savannah or a Temple Garden
// as readily as a basic. Myriad Landscape's activated shape with
// two searches in place of one: the prompt carries one clause, so
// the Forest search runs first and its continuation runs the Plains
// search, and the library is shuffled once after the second. Failing
// to find a Forest does not forfeit the Plains.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "d9a10971-f32b-4978-952d-fed0a5bc9e36",
		Name:          "Krosan Verge",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label:  "{2}, {T}, Sacrifice this land: Search your library for a Forest card and a Plains card, put them onto the battlefield tapped, then shuffle.",
			Cost:   Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: b08KrosanVergeFetch,
		}},
	})
}
