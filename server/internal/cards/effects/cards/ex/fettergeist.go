package cardsex

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fettergeist — {2}{U} Creature — Spirit 3/4: "Flying. At the beginning of your
// upkeep, sacrifice this creature unless you pay {1} for each other creature you
// control."
//
// The count is taken as the trigger resolves; with no other creature the payment
// is nothing and the Fettergeist stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ed73cff-137c-4ad1-b11d-c62aa746655e",
		Name:            "Fettergeist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Fettergeist — sacrifice this creature unless you pay {1} for each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					others := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Creature())) - 1
					if others <= 0 {
						return nil
					}
					cost := "{" + strconv.Itoa(others) + "}"
					return effects.UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     cost,
						Question: "Fettergeist — pay " + cost + " or sacrifice Fettergeist?",
						Else:     effects.UpkeepSacrificeSource,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
