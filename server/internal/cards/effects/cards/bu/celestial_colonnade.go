package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestial Colonnade — Land:
//
//	"This land enters tapped.
//	 {T}: Add {W} or {U}.
//	 {3}{W}{U}: Until end of turn, this land becomes a 4/4 white and blue
//	 Elemental creature with flying and vigilance. It's still a land."
//
// A manland built like Creeping Tar Pit: one continuous effect from the
// resolving ability, pinned to this object, with the type and subtype
// (layer 4), the colours (layer 5), the keywords (layer 6) and the base
// 4/4 (layer 7b). The engine supplies CR 302.6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "876ac6f6-74de-4666-84c6-83d81f054723",
		Name:         "Celestial Colonnade",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{W}{U}: Until end of turn, this land becomes a 4/4 white and blue Elemental creature with flying and vigilance. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{3}{W}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Celestial Colonnade — a 4/4 white and blue Elemental creature with flying and vigilance until end of turn",
					4, 4, "Elemental", []string{"W", "U"}, "flying", "vigilance")
			},
		}},
	})
}
