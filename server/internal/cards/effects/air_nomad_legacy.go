package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Air Nomad Legacy — Enchantment {W}{U} (Avatar: The Last Airbender,
// #4459):
//
//	"When this enchantment enters, create a Clue token.
//	 Creatures you control with flying get +1/+1."
//
// The anthem is Favorable Winds' (creaturesYouControlWithFlyingGetPlusOne),
// flying read through the layers. Creating the Clue is not investigating
// (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e830359-ab92-4bfd-82b3-d9c67ac80cf7",
		Name:         "Air Nomad Legacy",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Air Nomad Legacy — create a Clue", Do(CreateToken{Template: ClueToken(), N: 1})),
		},
		Static: []game.StaticAbility{creaturesYouControlWithFlyingGetPlusOne()},
	})
}
