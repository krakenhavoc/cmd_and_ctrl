package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loam Dweller — Creature — Spirit {1}{G}, 2/2 (Betrayers of Kamigawa,
// #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a land card
//	 from your hand onto the battlefield tapped."
//
// The Champions of Kamigawa cast trigger (chkB1SpiritOrArcane) with
// Arboreal Grazer's put: MayPutALandFromHandTapped asks as it resolves,
// "no" being a real answer, and with no land in hand nothing is asked.
// The land is put, not played, so it is not a land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a61f7653-4c55-472f-b37c-617b94d0fecd",
		Name:         "Loam Dweller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ChkB1SpiritOrArcane(),
				"Loam Dweller — you may put a land card from your hand onto the battlefield tapped",
				effects.Do(effects.MayPutALandFromHandTapped("Loam Dweller"))),
		},
	})
}
