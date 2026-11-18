package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hissing Quagmire — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B} or {G}.
//	 {1}{B}{G}: Until end of turn, this land becomes a 2/2 black and green
//	 Elemental creature with deathtouch. It's still a land."
//
// Celestial Colonnade's manland shape: type and subtype, colours,
// deathtouch and the base 2/2, all until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fff5224-c002-4a32-86fe-9a4b18d78b50",
		Name:         "Hissing Quagmire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{G}: Until end of turn, this land becomes a 2/2 black and green Elemental creature with deathtouch. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{1}{B}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return cmdr2B13Animate(NewContext(g, item), item.SourceCardID,
					"Hissing Quagmire — a 2/2 black and green Elemental creature with deathtouch until end of turn",
					2, 2, "Elemental", []string{"B", "G"}, "deathtouch")
			},
		}},
	})
}
