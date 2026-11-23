package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Quartermaster — Creature — Angel Soldier {3}{W}{W}, 3/3
// (Innistrad: Crimson Vow, #5495):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on each of up to two
//	 other target creatures."
//
// Basri's Acolyte's trigger without "you control": the targets are any
// creatures but the Quartermaster itself, zero, one or two of them, chosen
// as the trigger goes on the stack; each still legal as it resolves gets a
// counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "14a555eb-9720-45b4-9eb5-77b739e1045c",
		Name:            "Angelic Quartermaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Angelic Quartermaster — put a +1/+1 counter on each of up to two other target creatures",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.Another(effects.TargetCreature("up to two other target creatures")).WithCount(0, 2)),
		},
	})
}
