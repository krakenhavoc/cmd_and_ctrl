package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dramatic Finale — Enchantment {W/B}{W/B}{W/B}{W/B} (Strixhaven: School of
// Mages, #5502):
//
//	"Creature tokens you control get +1/+1.
//	 Whenever one or more nontoken creatures you control die, create a 2/1
//	 white and black Inkling creature token with flying. This ability triggers
//	 only once each turn."
//
// The static is Phantom General's layer 7c modify over the controller's creature
// tokens. The death trigger is Homicide Investigator's: a batch (CR 603.2c) of
// nontoken creatures that died under your control makes one Inkling, and
// TriggersOnlyOnceEachTurn spends the turn's one use.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0155236-13c5-4e01-b4de-b8fa4af846d8",
		Name:         "Dramatic Finale",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Creature tokens you control get +1/+1.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.B05CreatureTokenYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.On(game.EventLTB, effects.SkullsporeNontokenCreatureYouControlDied,
				"Dramatic Finale — create a 2/1 white and black Inkling with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 white and black Inkling with flying"), N: 1})))),
		},
	})
}
