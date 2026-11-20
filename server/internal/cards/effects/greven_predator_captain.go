package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greven, Predator Captain — Legendary Creature — Phyrexian Human Warrior
// {3}{B}{R}, 5/5 (Commander 2019, #5517):
//
//	"Menace
//	 Greven gets +X/+0, where X is the amount of life you've lost this
//	 turn.
//	 Whenever Greven attacks, you may sacrifice another creature. If you
//	 do, you draw cards equal to that creature's power and you lose life
//	 equal to that creature's toughness."
//
// The bonus is a layer 7c static read off the per-turn tally of life its
// controller has LOST (Children of Korlis's tally), so life gained back
// later in the turn does not shrink it. The attack trigger is Disciple of
// Freyalise's: the sacrifice is chosen as the trigger resolves, over the
// other creatures you control then (maySacrificeOne, CR 608.2c), and the
// payout follows only a creature that really left. Both numbers are that
// creature's last-known power and toughness (CR 608.2h), recorded once; the
// draw comes first and the life loss waits for any prompt the draw asks
// (a dredge offer). A negative number draws or loses nothing (CR 107.1b).
// The life lost this way raises Greven's power at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b22f4797-c74f-4d95-9528-c16c2bc4c05b",
		Name:            "Greven, Predator Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			Label:     "Greven gets +X/+0, where X is the amount of life you've lost this turn.",
			AppliesTo: selfOnly,
			// A life loss moves X with nothing on the battlefield changing.
			DependsOnLifeTotal: true,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += b18LifeLostThisTurn(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Greven, Predator Captain — you may sacrifice another creature; if you do, draw cards equal to its power and lose life equal to its toughness",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Greven, Predator Captain — sacrifice another creature to draw cards equal to its power and lose life equal to its toughness?",
						Creature(), true, c19GrevenSacrificed)
				}),
		},
	})
}

// c19GrevenSacrificedRecord files the sacrificed creature's power and
// toughness for the steps that pay them out.
const c19GrevenSacrificedRecord = "greven-sacrificed-pt"

// c19GrevenSacrificed is "if you do", once the creature has really gone.
// Keyed: an on-disk identity, never renamed or reused.
var c19GrevenSacrificed = RegisterSacrificedOne("prompt-run/greven-predator-captain-sacrifice", func(ctx *Context, sacrificed uuid.UUID) error {
	ctx.RecordStep(c19GrevenSacrificedRecord, nil,
		departedCreaturePower(ctx.Game, sacrificed), departedCreatureToughness(ctx.Game, sacrificed))
	return c19GrevenDrawThenLose.Run(ctx)
})

// c19GrevenDrawThenLose is "you draw cards equal to that creature's power
// and you lose life equal to that creature's toughness".
var c19GrevenDrawThenLose = RegisterThenSteps("then-steps/greven-predator-captain-draw-then-lose",
	Instruction(func(ctx *Context) error {
		if n := ctx.StepRecord(c19GrevenSacrificedRecord).N(); n > 0 {
			return DrawCards{N: n}.Apply(ctx)
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		ns := ctx.StepRecord(c19GrevenSacrificedRecord).Ns
		if len(ns) < 2 || ns[1] <= 0 {
			return nil
		}
		return GainLife{Amount: -ns[1]}.Apply(ctx)
	}),
)
