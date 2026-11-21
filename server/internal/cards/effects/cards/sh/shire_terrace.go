package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shire Terrace — Land (EDHREC rank 2284):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Search your library for a basic
//	 land card, put it onto the battlefield tapped, then shuffle."
//
// Evolving Wilds that taps for mana first. The mana ability is
// Buried Ruin's; the fetch is Evolving Wilds' activation with a {1}
// added to the cost, sharing its body (fetchBasicTapped): the
// searcher picks from the basic land cards, the pick enters tapped
// and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "619173f4-0403-49cd-9659-2fedd5028a90",
		Name:          "Shire Terrace",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.TapForColorless()},
		Activated:     []effects.ActivatedAbility{effects.OneTapSacFetchBasicTapped()},
	})
}
