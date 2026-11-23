package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vela the Night-Clad — Legendary Creature — Human Wizard {4}{U}{B}, 4/4
// (Commander Starter Decks):
//
//	"Intimidate
//	 Other creatures you control have intimidate.
//	 Whenever Vela or another creature you control leaves the battlefield,
//	 each opponent loses 1 life."
//
// Intimidate is the engine's block rule (CR 702.13b); the grant is Heliod's
// "other creatures you control" as a layer 6 keyword grant. The trigger reads
// the leaver's controller from last-known information (CR 603.10), so a
// creature you control that dies, is exiled or is bounced counts, Vela
// included, and each one that leaves is its own trigger.
//
// No simplification.
func init() {
	grant := effects.KeywordGrant(effects.OtherCreaturesYouControl, "intimidate")
	grant.Label = "Other creatures you control have intimidate."
	effects.Register(effects.Spec{
		OracleID:        "185e34f0-60ae-4b07-94ac-96cef11b1604",
		Name:            "Vela the Night-Clad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Static:          []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.ACreatureYouControlLeft,
				"Vela the Night-Clad — each opponent loses 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
		},
	})
}
