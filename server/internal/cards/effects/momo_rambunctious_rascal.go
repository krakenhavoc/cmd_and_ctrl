package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Momo, Rambunctious Rascal — Legendary Creature — Lemur Bat Ally {2}{W},
// 1/1 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Flying
//	 When Momo enters, he deals 4 damage to target tapped creature an
//	 opponent controls."
//
// Shuttle Crew's enters trigger, word for word
// (whenThisEntersFourDamageToTappedOpponentCreature).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8854222e-590f-4fc0-b5d7-bf3d986b8507",
		Name:            "Momo, Rambunctious Rascal",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			whenThisEntersFourDamageToTappedOpponentCreature("Momo, Rambunctious Rascal — he deals 4 damage to target tapped creature an opponent controls"),
		},
	})
}
