package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verix Bladewing — Legendary Creature — Dragon {2}{R}{R}, 4/4 (Tarkir:
// Dragonstorm Commander):
//
//	"Kicker {3}
//	 Flying
//	 When Verix Bladewing enters, if it was kicked, create Karox
//	 Bladewing, a legendary 4/4 red Dragon creature token with flying."
//
// Archangel of Wrath's kicker and intervening-if enters trigger
// (ThisKickedAtLeast, read off the record the resolution carried onto the
// permanent, CR 603.4 / 400.7d): a Verix that is put onto the battlefield
// without being cast isn't kicked and makes nothing. The token is the
// printed Karox Bladewing, legendary, named, with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d7315b3b-6e5d-4d28-a503-bfae1daf6c43",
		Name:            "Verix Bladewing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{effects.Kicker("{3}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
				"Verix Bladewing — create Karox Bladewing, a legendary 4/4 red Dragon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Legendary Dragon named Karox Bladewing with flying"), N: 1})),
		},
	})
}
