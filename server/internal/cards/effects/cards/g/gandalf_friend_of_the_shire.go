package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gandalf, Friend of the Shire — Legendary Creature — Avatar Wizard
// {3}{U}, 2/4:
//
//	"Flash
//	 You may cast sorcery spells as though they had flash.
//	 Whenever the Ring tempts you, if you chose a creature other than
//	 Gandalf as your Ring-bearer, draw a card."
//
// The second ability is Teferi, Time Raveler's +1 as a static: it
// changes when you may cast sorceries, not when you may activate an
// ability "only as a sorcery" (2023-06-16 ruling).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7f7c917b-f940-4f43-bafa-31f50a6d9d6f",
		Name:            "Gandalf, Friend of the Shire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CastTimings: []game.CastTimingRule{
			effects.CastKindAsThoughFlash(game.PermissionFilter{SorceryOnly: true}, "You may cast sorcery spells as though they had flash."),
		},
		Triggered: []game.TriggeredAbility{
			effects.IfYouChoseAnotherRingBearer(effects.WheneverTheRingTemptsYou("Gandalf, Friend of the Shire — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
