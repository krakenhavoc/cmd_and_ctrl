package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thelonite Monk — Creature — Insect Monk Cleric {2}{G}{G}, 1/2:
//
//	"{T}, Sacrifice a green creature: Target land becomes a Forest. (This
//	 effect lasts indefinitely.)"
//
// ADR 0109 §1 decision 5 (#1881): CR 305.7's type set with no duration, so
// it lasts until the game ends (CR 611.2a), pinned to the land so it ends
// if the land stops being that object (CR 400.7). The land's land types
// are replaced by Forest (other subtypes stay, CR 205.1a), it loses its
// rules-text abilities and taps for {G} (CR 305.6). The Monk is green and
// may sacrifice itself, since {T} is paid first and the sacrifice does not
// need it untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26b0b742-48f7-4ab5-93de-6bd39a7f61ea",
		Name:         "Thelonite Monk",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice a green creature: Target land becomes a Forest. (This effect lasts indefinitely.)",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "a green creature", effects.Creature(), effects.OfColor("G"))),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.TargetLandBecomesIndefinitely("Thelonite Monk", "Forest"),
		}},
	})
}
