package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Threnody Singer — Creature — Siren {1}{U}, 1/3 (Theros Beyond Death, #5514):
//
//	"Flash, flying. When this creature enters, target creature an opponent controls
//	 gets -X/-0 until end of turn, where X is your devotion to blue."
//
// Devotion is counted as the trigger resolves (CR 700.5, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c0816bd-8ce2-4ddf-845a-bcdaf929fe28",
		Name:            "Threnody Singer",
		PrintedKeywords: []string{"flash", "flying"},
		Completeness:    effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Threnody Singer — target creature an opponent controls gets -X/-0 until end of turn",
				effects.S82OldB133DevotionBoostFirstTarget("Threnody Singer — -X/-0", "U", -1, 0)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, 0, 0))),
		},
	})
}
