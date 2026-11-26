package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Medomai the Ageless — Legendary Creature — Sphinx {4}{W}{U}, 4/4 (Theros):
//
//	"Flying
//	 Whenever Medomai deals combat damage to a player, take an extra turn
//	 after this one.
//	 Medomai can't attack during extra turns."
//
// The trigger is Wormfang Manta's extra turn (CR 500.7). The restriction is an
// attack-target restriction read live at declaration: during any turn an
// effect made (Turn.Extra) Medomai may not be declared as an attacker.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6b91ea7-a22c-4835-8c62-3360b0fd3a8b",
		Name:            "Medomai the Ageless",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{ExtraTurns: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Medomai the Ageless — take an extra turn after this one", effects.Do(effects.TakeExtraTurn{})),
		},
		Static: []game.StaticAbility{effects.OldB394CantAttackDuringExtraTurns()},
	})
}
