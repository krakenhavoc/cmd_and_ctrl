package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Roc Egg — Creature — Bird Egg {2}{W}, 0/3 (Magic 2011, #5579):
//
//	"Defender
//	 When this creature dies, create a 3/3 white Bird creature token with
//	 flying."
//
// Defender is the printed keyword. The dies trigger is Dwarven Castle
// Guard's token trigger with the 3/3 flying Bird, made for the Egg's
// controller as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "16dfd80a-9847-45b8-954f-2af9b7cfa21d",
		Name:            "Roc Egg",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Roc Egg — create a 3/3 white Bird creature token with flying",
				Do(CreateToken{Template: TokenCard("3/3 white Bird with flying"), N: 1})),
		},
	})
}
