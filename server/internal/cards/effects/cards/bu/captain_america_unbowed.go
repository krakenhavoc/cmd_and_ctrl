package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "4fa2c9cb-106c-482c-9ba7-5bba2ea9e4a3",
		Name:            "Captain America, Unbowed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Captain America, Unbowed — Soldiers and Heroes you control gain indestructible until end of turn",
				effects.Do(effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl(), effects.AnySubtype("Soldier", "Hero")),
					Keywords: []string{"indestructible"},
					Label:    "Captain America, Unbowed",
				})),
		},
	})
}
