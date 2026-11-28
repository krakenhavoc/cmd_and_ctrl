package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Dragon — Creature — Dragon {6}{R}{R}{R}, 6/6 (Portal, S82 OLD batch 782,
// #5630):
//
//	"Flying
//	 When this creature enters, it deals damage to target creature equal to the
//	 number of Mountains you control."
//
// Cat-Gator's count on an entry trigger, aimed at a creature. The Mountains are
// counted once, as the trigger resolves (CR 608.2h), by effective land subtype;
// the Dragon is the source of the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "349d8ef9-e07a-416f-a5b1-2c2be6bb322d",
		Name:            "Fire Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Fire Dragon — damage to target creature equal to the number of Mountains you control",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					return effects.CountControlled(ctx.Game, ctx.Controller(), effects.MatchLandSubtype(ctx.TextSubtype("Mountain")))
				})), effects.TargetCreature("target creature")),
		},
	})
}
