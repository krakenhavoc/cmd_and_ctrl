package effects

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lazotep Convert — Creature — Zombie, the back face of Invasion of
// Amonkhet (March of the Machine, S82 batch 23):
//
//	"You may have this creature enter as a copy of any creature card in a
//	 graveyard, except it's a 4/4 black Zombie in addition to its other
//	 colors and types."
//
// Body Double's replacement (CR 707.2, CR 614.1c) with a different except
// clause: the copy takes the card's printed values, then is 4/4 (replacing
// the copied power and toughness) and black and a Zombie in addition to
// its other colours and creature types. Cast transformed, the back face is
// the spell, so it enters as that copy (or as its printed 4/4 Zombie when
// declined: a Siege's back face keeps the face's own values).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     invasionOfAmonkhetOracleID + "#1",
		Name:         "Lazotep Convert",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf("Lazotep Convert", anyCreatureCardInAGraveyard, lazotepConvertExcept),
		},
	})
}

// lazotepConvertExcept is the except clause: a 4/4 black Zombie in addition
// to its other colours and types.
func lazotepConvertExcept(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
	v.SetPowerToughness(4, 4)
	v.AddSubtype("Zombie")
	if !slices.Contains(v.Colors, "B") {
		v.Colors = append(slices.Clone(v.Colors), "B")
	}
}
