package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foot Mystic — {3}{B} Creature — Human Ninja Warlock 2/4 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Lifelink
//	 Disappear — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, create a 1/1 black Ninja
//	 creature token."
//
// Disappear is an ability word (CR 207.2c) for revolt's condition, so the
// trigger is WhenThisEntersIfRevolt: an intervening "if" checked as it
// triggers and again as it resolves (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2043fbde-48c4-4a77-8911-8991d77de1eb",
		Name:            "Foot Mystic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersIfRevolt("Foot Mystic — disappear: create a 1/1 black Ninja creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 black Ninja"), N: 1})),
		},
	})
}
