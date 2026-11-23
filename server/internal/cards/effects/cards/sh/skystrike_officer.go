package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skystrike Officer — Creature — Human Soldier {2}{U}, 2/3 (The Brothers'
// War):
//
//	"Flying
//	 Whenever this creature attacks, create a 1/1 colorless Soldier
//	 artifact creature token.
//	 Tap three untapped Soldiers you control: Draw a card."
//
// The cost is Aphetto Grifter's tap-others cost with a count of three: it is
// not the {T} symbol (CR 302.6), so a Soldier that arrived this turn pays it,
// the Officer may be one of the three, and it does not target. The token is
// Third Path Iconoclast's Soldier artifact creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "862798c6-1c38-4bb9-b4a1-3ef1fcf8f7f7",
		Name:            "Skystrike Officer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Skystrike Officer — create a 1/1 colorless Soldier artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Soldier artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap three untapped Soldiers you control: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent("three untapped Soldiers you control", effects.Creature(), effects.OfCreatureType("Soldier")),
				Label:  "three untapped Soldiers you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
