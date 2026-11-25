package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blight-Breath Catoblepas — Creature — Beast {4}{B}{B}, 3/2 (Theros Beyond Death, #5514):
//
//	"When this creature enters, target creature an opponent controls gets -X/-X
//	 until end of turn, where X is your devotion to black."
//
// Devotion is counted as the trigger resolves (CR 700.5, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4098ab29-bf90-4ec9-b547-05d304fc7798",
		Name:         "Blight-Breath Catoblepas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Blight-Breath Catoblepas — target creature an opponent controls gets -X/-X until end of turn",
				effects.S82OldB133DevotionBoostFirstTarget("Blight-Breath Catoblepas — -X/-X", "B", -1, -1)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
				effects.ForTargets(effects.HarmUntilEOT(0, 0, 0))),
		},
	})
}
