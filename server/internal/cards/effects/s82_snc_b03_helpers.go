package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_snc_b03_helpers.go — the shared piece of Streets of New Capenna
// batch 3 (#5491).

// sncCyclingTapland is the Streets of New Capenna sacrifice-land cycle
// (Botanical Plaza, Racers' Ring, Skybridge Towers, Tramway Station,
// Waterfront District):
//
//	"This land enters tapped.
//	 {T}: Add {A} or {B}.
//	 {2}{A}{B}, {T}, Sacrifice this land: Draw a card."
//
// It is avatarB20CyclingTapland's shape with the set's cost: a CR 602
// ability whose cost sacrifices the land, so the land is gone before the
// card is drawn and a response can't save it.
func sncCyclingTapland(oracleID, name, a, b string) Spec {
	cost := "{2}{" + a + "}{" + b + "}"
	return Spec{
		OracleID:      oracleID,
		Name:          name,
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility(a, b)},
		Activated: []ActivatedAbility{{
			Label:   cost + ", {T}, Sacrifice this land: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost(cost), TapCost(), SacrificeThis()),
			Effect:  drawOneCardEffect,
		}},
	}
}
