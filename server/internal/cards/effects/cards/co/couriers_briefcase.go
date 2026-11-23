package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Courier's Briefcase — Artifact — Treasure {1}{G} (Streets of New Capenna):
//
//	"When this artifact enters, create a 1/1 green and white Citizen
//	 creature token.
//	 {T}, Sacrifice this artifact: Add one mana of any color.
//	 {W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Draw three cards."
//
// A Treasure card (the subtype is printed, so Treasure payoffs see it) with
// the Treasure mana ability, Glittering Stockpile's shape. The draw ability
// is an ordinary activated ability whose cost sacrifices the source, so it
// resolves even though the Briefcase is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00cd3bf9-0dd6-4ec0-a841-1839a71604ab",
		Name:         "Courier's Briefcase",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Courier's Briefcase — create a 1/1 green and white Citizen creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}, Sacrifice this artifact: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Cost:    effects.Plus(effects.ManaCost("{W}{U}{B}{R}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
