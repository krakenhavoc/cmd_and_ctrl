package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Katara, Heroic Healer — Legendary Creature — Human Warrior Ally {4}{W},
// 2/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Lifelink
//	 When Katara enters, put a +1/+1 counter on each other creature you
//	 control."
//
// Dawnstrike Vanguard's body (dawnstrikeCounters): every other creature
// its controller controls as the trigger resolves, one placement each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "703032eb-98b3-4ecf-8b92-4f4f1551c5d7",
		Name:            "Katara, Heroic Healer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Katara, Heroic Healer — put a +1/+1 counter on each other creature you control", dawnstrikeCounters),
		},
	})
}
