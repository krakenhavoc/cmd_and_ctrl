package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dungeon Descent — Land:
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 {4}, {T}, Tap an untapped legendary creature you control: Venture
//	 into the dungeon. Activate only as a sorcery."
//
// The enters-tapped is a self-replacement (SelfEntersTapped), and
// tapping the legendary creature is a cost that is not the {T} symbol,
// so one that arrived this turn may pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f086a63c-0c62-4674-bd27-82e7aed12b1a",
		Name:         "Dungeon Descent",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{4}, {T}, Tap an untapped legendary creature you control: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{4}"), effects.TapCost(), effects.TapAnUntapped("an untapped legendary creature you control", effects.Creature(), effects.Legendary())),
			SorcerySpeed: true,
			Effect:       effects.Venture,
		}},
	})
}
