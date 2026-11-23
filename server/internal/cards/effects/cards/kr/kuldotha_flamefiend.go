package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kuldotha Flamefiend — Creature — Elemental {4}{R}{R}, 4/4 (Jumpstart 2022,
// #5948):
//
//	"When this creature enters, you may sacrifice an artifact. If you do,
//	 this creature deals 4 damage divided as you choose among any number of
//	 targets."
//
// A targeted enters trigger: the targets (any number, of any kind) and how the
// 4 damage is divided are announced as it goes on the stack (CR 601.2d, Fury's
// fixed-4 shape), and the sacrifice is asked as it resolves (CR 608.2c, Boilerbilges
// Ripper's "you may sacrifice"). Only an artifact that really went raises the
// "if you do": the damage is then dealt by the Flamefiend, as it last existed if
// it has left (CR 608.2h), to each target still legal. With no artifact to
// sacrifice nobody is asked and nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e93089b1-e790-402b-a994-69a7e422e4f2",
		Name:         "Kuldotha Flamefiend",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Kuldotha Flamefiend — you may sacrifice an artifact; if you do, 4 damage divided among any number of targets",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Kuldotha Flamefiend — sacrifice an artifact to deal 4 damage divided as you chose?",
						effects.Artifact(), false, kuldothaFlamefiendSacrificed)
				}),
				effects.TargetAny().WithCount(0, 0).Dividing(effects.Divide(4))),
		},
	})
}

// kuldothaFlamefiendSacrificed is the "if you do": the divided damage, keyed.
var kuldothaFlamefiendSacrificed = effects.RegisterSacrificedOne("prompt-run/s82-gap-b156-kuldotha-flamefiend-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.DealDividedDamage(ctx)
})
