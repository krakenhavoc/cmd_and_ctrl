package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "72e0f7f2-1d04-4ea8-a6e8-4d6a13aae9c7",
		Name:         "Desperate Sentry",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +3/+0 as long as there are four or more card types among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(3, 0),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Desperate Sentry — create a 3/2 colorless Eldrazi Horror creature token",
				Do(CreateToken{Template: TokenCard("3/2 colorless Eldrazi Horror"), N: 1})),
		},
	})
}
