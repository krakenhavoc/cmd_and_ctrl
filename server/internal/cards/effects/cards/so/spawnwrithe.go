package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spawnwrithe — Creature — Elemental {2}{G}, 2/2 (Commander 2011, #5574):
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, create a token
//	 that's a copy of this creature."
//
// The copy is a token copy of the Spawnwrithe as it is when the ability
// resolves, or as it last existed on the battlefield (CR 707.2, 608.2h). Each
// copy has the trigger too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "afc59d95-d85c-4ab8-b6c5-c78b608b53b7",
		Name:            "Spawnwrithe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Spawnwrithe — create a token that's a copy of this creature",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: ctx.Source(), N: 1}.Apply(ctx)
				}))),
		},
	})
}
