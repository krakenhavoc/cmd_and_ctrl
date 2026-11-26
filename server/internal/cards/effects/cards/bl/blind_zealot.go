package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blind Zealot — Creature — Phyrexian Human Cleric {1}{B}{B}, 2/2:
//
//	"Intimidate
//	 Whenever this creature deals combat damage to a player, you may sacrifice
//	 it. If you do, destroy target creature that player controls."
//
// The target is chosen as the trigger goes on the stack, among the creatures
// the damaged player controls; the destruction happens only if the Zealot
// really was sacrificed (it is not, if it had already left, CR 400.7).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30c7e304-ecfe-44c1-88d1-f507df63c080",
		Name:            "Blind Zealot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Blind Zealot — you may sacrifice it; if you do, destroy target creature that player controls",
				effects.SacrificeThisThen("then/s82-old-b461-blind-zealot-destroy", effects.DestroyTheFirstLegalCardTarget)),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
				"Blind Zealot — sacrifice it to destroy target creature that player controls?"),
		},
	})
}
