package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broodmate Tyrant — Creature — Dragon {4}{B}{R}{G}, 5/5 (Modern Horizons
// 3 Commander, #5468):
//
//	"Flying
//	 When this creature enters, create a 5/5 red Dragon creature token
//	 with flying.
//	 Encore {5}{B}{R}{G}"
//
// The enters trigger makes one Dragon for its controller. Encore is the
// shared constructor (one hasty token copy per opponent, each bound to
// attack that opponent, sacrificed at the next end step); each copy is a
// Broodmate Tyrant, so each one's enters trigger makes a Dragon too, and
// those Dragons stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b984236-39cf-4552-827b-e81c26cfb388",
		Name:            "Broodmate Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Broodmate Tyrant — create a 5/5 red Dragon with flying",
				Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{Encore("{5}{B}{R}{G}")},
	})
}
