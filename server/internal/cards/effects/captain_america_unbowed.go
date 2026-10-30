package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captain America, Unbowed — Legendary Creature — Human Soldier Hero {3}{W},
// 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 When Captain America enters, Soldiers and Heroes you control gain
//	 indestructible until end of turn."
//
// Flawless Maneuver's grant with a Soldier-or-Hero filter, chosen when the
// trigger resolves: a Soldier or Hero that arrives afterwards gets nothing.
// Captain America is both, so he covers himself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4fa2c9cb-106c-482c-9ba7-5bba2ea9e4a3",
		Name:            "Captain America, Unbowed",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Captain America, Unbowed — Soldiers and Heroes you control gain indestructible until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl(), AnySubtype("Soldier", "Hero")),
					Keywords: []string{"indestructible"},
					Label:    "Captain America, Unbowed",
				})),
		},
	})
}
