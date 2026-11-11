package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hostile Desert — Land — Desert:
//
//	"{T}: Add {C}.
//	 {2}, Exile a land card from your graveyard: This land becomes a 3/4
//	 Elemental creature until end of turn. It's still a land."
//
// The exile is a cost (CR 118.1). The animation sets no colour, so the
// creature is colourless, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "41459587-7509-404e-bd7d-fb8831dee789",
		Name:         "Hostile Desert",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Exile a land card from your graveyard: This land becomes a 3/4 Elemental creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    Plus(ManaCost("{2}"), ExileFromGraveyard(1, "a land card", isLandCard)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return cmdr2B13Animate(NewContext(g, item), item.SourceCardID,
					"Hostile Desert — a 3/4 Elemental creature until end of turn", 3, 4, "Elemental", nil)
			},
		}},
	})
}
