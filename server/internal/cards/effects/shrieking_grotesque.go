package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "20c8b03c-303d-4e19-96ec-3fabae843352",
		Name:            "Shrieking Grotesque",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ReadsManaSpent:  game.ReadsOwnPayment,
		WantsColors:     []string{"B"},
		Triggered: []game.TriggeredAbility{
			gptWhenEntersIfSpent("B", "Shrieking Grotesque — target player discards a card",
				TargetPlayer("target player"), kemuriOnnaDiscard),
		},
	})
}
