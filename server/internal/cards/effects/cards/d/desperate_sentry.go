package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Sentry — Creature — Human Soldier {2}{W}, 1/2 (Eldritch Moon,
// #5541):
//
//	"When this creature dies, create a 3/2 colorless Eldrazi Horror
//	 creature token.
//	 Delirium — This creature gets +3/+0 as long as there are four or more
//	 card types among cards in your graveyard."
//
// Doomed Dissenter's dies trigger with Howling Chorus's token row, and
// Gnarlwood Dryad's delirium static at +3/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72e0f7f2-1d04-4ea8-a6e8-4d6a13aae9c7",
		Name:         "Desperate Sentry",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +3/+0 as long as there are four or more card types among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature),
			Apply:     effects.PumpBy(3, 0),
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Desperate Sentry — create a 3/2 colorless Eldrazi Horror creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 colorless Eldrazi Horror"), N: 1})),
		},
	})
}
