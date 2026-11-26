package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krond the Dawn-Clad — Legendary Creature — Archon {G}{G}{G}{W}{W}{W}, 6/6
// (Planechase 2012, S82 OLD, #5569):
//
//	"Flying, vigilance
//	 Whenever Krond attacks, if it's enchanted, exile target permanent."
//
// An attack trigger with an intervening "if" (CR 603.4): an Aura is attached to
// Krond as it attacks and again as the trigger resolves, so removing it in
// response fizzles the exile. The permanent is targeted as the trigger is put on
// the stack and exiled if still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "62ca69ed-ebd3-4e84-b17a-4bcc9488f8b3",
		Name:            "Krond the Dawn-Clad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.S82OldB432InterveningIfEnchanted(
				effects.WheneverThisAttacks("Krond the Dawn-Clad — exile target permanent", effects.Do(effects.ExileChosenTarget{}))),
				effects.TargetPermanent("target permanent")),
		},
	})
}
