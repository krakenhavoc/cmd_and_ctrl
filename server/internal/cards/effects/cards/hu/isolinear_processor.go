package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Isolinear Processor — Artifact {2} (Star Trek, #4973):
//
//	"{T}: Add {C}. Spend this mana only to cast an artifact spell or
//	 activate an ability of an artifact source.
//	 {2}, {T}: Draw a card, then discard a card."
//
// The mana carries Gwenna, Eyes of Gaea's restriction shape with
// "Artifact" for "Creature": casting an artifact spell, or activating an
// ability whose source is an artifact (CR 106.6). Any other purpose is a
// refusal. The loot is Bag of Holding's (b16DrawThenDiscard).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef5d7b38-b782-4fb0-b6cd-3da36995ab39",
		Name:         "Isolinear Processor",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
