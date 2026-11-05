package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ef5d7b38-b782-4fb0-b6cd-3da36995ab39",
		Name:         "Isolinear Processor",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "{T}: Add {C}. Spend this mana only to cast an artifact spell or activate an ability of an artifact source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictType("Artifact")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictType("Artifact")},
			)},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card, then discard a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 1)
			},
		}},
	})
}
