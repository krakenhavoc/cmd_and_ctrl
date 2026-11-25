package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Vial — Artifact {3} (Zendikar Rising, S82 OLD #5508):
//
//	"{2}, {T}, Sacrifice a creature: Draw a card.
//	 As long as you control a Cleric, this artifact has "Whenever a
//	 creature you control dies, each opponent loses 1 life and you gain 1
//	 life.""
//
// The second line is a layer 6 grant of a triggered ability to the Vial
// itself while its controller controls a Cleric (a changeling counts, CR
// 702.73a). A creature dying together with the last Cleric still triggers
// it: leaves-the-battlefield abilities look back to the battlefield as the
// event left it (CR 603.10a).
//
// No simplification.
func init() {
	const grant = "relic-vial/drain"
	effects.Register(effects.Spec{
		OracleID:     "7e86d4a6-e4cd-418a-8bbc-72cdad753e3a",
		Name:         "Relic Vial",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverACreatureYouControlDies("Relic Vial — each opponent loses 1 life and you gain 1 life",
					effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
			},
			Text: "Whenever a creature you control dies, each opponent loses 1 life and you gain 1 life.",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && effects.ControlsPartyType(g, source.Controller, "Cleric")
			}, grant), "As long as you control a Cleric, this artifact has \"Whenever a creature you control dies, each opponent loses 1 life and you gain 1 life.\""),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice a creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
