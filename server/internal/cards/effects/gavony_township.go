package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gavony Township — Land (EDHREC rank 999):
//
//	"{T}: Add {C}.
//	 {2}{G}{W}, {T}: Put a +1/+1 counter on each creature you control."
//
// The Selesnya token deck's mana sink: every spare four mana is an
// anthem that stays. A colorless mana ability and one activated
// ability whose effect is Cathars' Crusade's body — the set of
// creatures you control is snapshotted before the first counter is
// placed, so a creature that arrives mid-resolution gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8a44e4e7-dfa2-427b-bbff-11c398fa60bb",
		Name:          "Gavony Township",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label: "{2}{G}{W}, {T}: Put a +1/+1 counter on each creature you control.",
			Cost:  Plus(ManaCost("{2}{G}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b11PutCountersOnEachCreatureYouControl(g, item, 1)
			},
		}},
	})
}
