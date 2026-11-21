package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "41459587-7509-404e-bd7d-fb8831dee789",
		Name:         "Hostile Desert",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Exile a land card from your graveyard: This land becomes a 3/4 Elemental creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileFromGraveyard(1, "a land card", effects.IsLandCard)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.Cmdr2B13Animate(effects.NewContext(g, item), item.SourceCardID,
					"Hostile Desert — a 3/4 Elemental creature until end of turn", 3, 4, "Elemental", nil)
			},
		}},
	})
}
