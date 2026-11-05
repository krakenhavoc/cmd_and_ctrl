package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b30_helpers.go — shared pieces for Star Trek batch 30 (#4973).

// tapForColorless is the plain "{T}: Add {C}." mana ability.
func tapForColorless() ManaAbility {
	return ManaAbility{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "Add {C}"}
}

// oneTapSacFetchBasicTapped is "{1}, {T}, Sacrifice this land: Search
// your library for a basic land card, put it onto the battlefield
// tapped, then shuffle." — Shire Terrace and Unexplored Territory, the
// Evolving Wilds body (fetchBasicTapped) behind a {1}.
func oneTapSacFetchBasicTapped() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{1}, {T}, Sacrifice this land: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
		Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
		Effect:  fetchBasicTapped,
	}
}
