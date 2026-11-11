package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Flayer — Creature — Human Warrior {B}{G}, 2/2 (Duskmourn Commander):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, surveil 3.
//	 Delirium — This creature gets +2/+2 as long as there are four or more
//	 card types among cards in your graveyard."
//
// The trigger is one surveil 3 per instance of combat damage to a player
// (CR 120.2); the bonus is Gnarlwood Dryad's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "21254bdb-7b1f-4ac6-8fdd-f0ac2c35f47d",
		Name:            "Grim Flayer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Grim Flayer — surveil 3", Do(Surveil{N: 3})),
		},
		Static: []game.StaticAbility{{
			Label:     "Delirium — Gets +2/+2 as long as there are four or more card types among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(2, 2),
		}},
	})
}
