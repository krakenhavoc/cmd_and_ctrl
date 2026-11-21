package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrieking Grotesque — Creature — Gargoyle {2}{W}, 2/1 (Guildpact,
// #5599):
//
//	"Flying
//	 When this creature enters, if {B} was spent to cast it, target
//	 player discards a card."
//
// Gruul Scrapper's intervening if (gptWhenEntersIfSpent) on Kemuri-Onna's
// discard (kemuriOnnaDiscard, the target's own choice of card). With no
// black mana spent nothing triggers. ReadsManaSpent and WantsColors as on
// the Scrapper.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "20c8b03c-303d-4e19-96ec-3fabae843352",
		Name:            "Shrieking Grotesque",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"B"},
		Triggered: []game.TriggeredAbility{
			effects.GptWhenEntersIfSpent("B", "Shrieking Grotesque — target player discards a card",
				effects.TargetPlayer("target player"), effects.KemuriOnnaDiscard),
		},
	})
}
