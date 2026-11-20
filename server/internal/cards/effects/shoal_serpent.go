package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shoal Serpent — Creature — Serpent {5}{U}, 5/5 (Zendikar, #5583):
//
//	"Defender
//	 Landfall — Whenever a land you control enters, this creature loses
//	 defender until end of turn."
//
// Defender is the printed keyword; landfall removes it until end of
// turn with Jaded Analyst's layer-6 removal (CR 613.1f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8a1d14c2-ca31-4711-b385-f4af0c1c0b04",
		Name:            "Shoal Serpent",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			Landfall("Shoal Serpent — loses defender until end of turn (landfall)",
				thisLosesKeywordUntilEndOfTurn("defender", "Shoal Serpent — loses defender until end of turn")),
		},
	})
}
