package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moldgraf Scavenger — Creature — Fungus {1}{G}, 0/4
// (Shadows over Innistrad, #5543):
//
//	"Delirium — This creature gets +3/+0 as long as there are four or
//	 more card types among cards in your graveyard."
//
// Delirium is an ability word (CR 207.2c): a layer 7c pump on this
// creature alone, gated on its controller's own graveyard (drcDeliriumMet)
// and re-read on every recompute, as on Grim Flayer.
//
// No simplification.
func init() {
	const delirium = " as long as there are four or more card types among cards in your graveyard."
	pump := game.StaticAbility{
		Label:     "Delirium — Gets +3/+0" + delirium,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: effects.WhileYouHave(effects.DrcDeliriumMet, effects.SelfCreature),
		Apply:     effects.PumpBy(3, 0),
	}
	effects.Register(effects.Spec{
		OracleID:     "558fe476-edbd-4f18-86ee-6d7f61331789",
		Name:         "Moldgraf Scavenger",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{pump},
	})
}
