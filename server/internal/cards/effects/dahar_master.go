package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dahar Master — {3}{R} Creature — Klingon Warrior, 4/3 (Star Trek,
// #4973):
//
//	"When this creature enters, you may sacrifice another nonland
//	 permanent. When you do, this creature deals 2 damage to any target."
//
// Bolg of the North's shape: the enters trigger asks "you may sacrifice
// another nonland permanent" as it resolves (maySacrificeOne, CR 608.2c),
// and only a permanent that really went raises the CR 603.12 reflexive
// trigger, which picks its target as it goes on the stack (CR 603.3d).
// The damage is Bullseye's (Params.Amount from the source to the chosen
// target); a Dahar Master that has left deals it as it last existed (CR
// 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9bc1a28-8d3c-4d9f-a67d-b044cfe9d6c1",
		Name:         "Dahar Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dahar Master — you may sacrifice another nonland permanent; when you do, 2 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Dahar Master — sacrifice another nonland permanent to deal 2 damage to any target?",
						Nonland(), true, daharMasterSacrificed)
				}),
		},
	})
}

// daharMasterSacrificed raises the reflexive trigger once a permanent has
// really gone, keyed.
var daharMasterSacrificed = RegisterSacrificedOne("prompt-run/dahar-master-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label:  "Dahar Master — 2 damage to any target",
		Body:   daharMasterDamageBody,
		Params: game.EffectParams{Amount: 2},
	}.Apply(ctx)
})

// daharMasterDamageBody is the reflexive trigger's registration: "deals 2
// damage to any target", the amount on Params.Amount.
var daharMasterDamageBody = game.ReflexiveBody("dahar-master/damage", breechesBlastEffect, constTargets(TargetAny))
