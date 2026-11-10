package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boilerbilges Ripper — Creature — Human Assassin {4}{R}, 4/4 (Duskmourn,
// #4466):
//
//	"When this creature enters, you may sacrifice another creature or
//	 enchantment. When you do, this creature deals 2 damage to any target."
//
// Dahar Master's shape: the enters trigger asks "you may sacrifice another
// creature or enchantment" as it resolves (maySacrificeOne, CR 608.2c), and
// only a permanent that really went raises the CR 603.12 reflexive trigger,
// which chooses its "any target" as it goes on the stack (the ruling), not
// when the parent was announced. The damage is the Ripper's, as it last
// existed if it has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cbe2a00-4303-4ade-98b5-4172d97f15fb",
		Name:         "Boilerbilges Ripper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Boilerbilges Ripper — you may sacrifice another creature or enchantment; when you do, 2 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Boilerbilges Ripper — sacrifice another creature or enchantment to deal 2 damage to any target?",
						Or(Creature(), Enchantment()), true, boilerbilgesRipperSacrificed)
				}),
		},
	})
}

// boilerbilgesRipperSacrificed raises the reflexive trigger once a
// permanent has really gone, keyed.
var boilerbilgesRipperSacrificed = RegisterSacrificedOne("prompt-run/boilerbilges-ripper-sacrifice", func(ctx *Context, _ uuid.UUID) error {
	return ReflexiveTrigger{
		Label:  "Boilerbilges Ripper — 2 damage to any target",
		Body:   boilerbilgesRipperDamageBody,
		Params: game.EffectParams{Amount: 2},
	}.Apply(ctx)
})

// boilerbilgesRipperDamageBody is the reflexive trigger's registration:
// "deals 2 damage to any target", the amount on Params.Amount.
var boilerbilgesRipperDamageBody = game.ReflexiveBody("boilerbilges-ripper/damage", breechesBlastEffect, constTargets(TargetAny))
