package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myrkul's Edict — Sorcery {1}{B} (Commander Legends: Battle for Baldur's
// Gate, S82 GAP #5948):
//
//	"Roll a d20.
//	 1—9 | Choose an opponent. That player sacrifices a creature of their
//	 choice.
//	 10—19 | Each opponent sacrifices a creature of their choice.
//	 20 | Each opponent sacrifices a creature with the greatest power among
//	 creatures that player controls."
//
// The d20 goes through the one dice door and its tail reads the table. 1—9
// asks the caster for an opponent (ChoosePlayer, Intellectual Offering's
// pick), then that player alone sacrifices; 10—19 is every opponent at once;
// 20 limits each opponent's choice to the creatures tied for the greatest
// power they control. Each sacrifice is a run of the ordered collector (Shadowgrange
// Archfiend's), so the seats are asked in APNAP order and everything chosen
// leaves together; an opponent with no creature sacrifices nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "933afaa3-3b23-454c-9465-8f2ba9652b4f",
		Name:         "Myrkul's Edict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RollDice{Sides: 20, N: 1, Then: myrkulsEdictRolled}.Apply(ctx)
		},
	})
}

var myrkulsEdictDone = effects.RunThen("prompt-run/myrkuls-edict-done", func(*effects.Context, game.RunDone) error { return nil })

// myrkulsEdictRolled reads the table. An on-disk key.
var myrkulsEdictRolled = effects.DiceTail("dice/myrkuls-edict-rolled", func(ctx *effects.Context, rolls []int) error {
	opponents := ctx.Opponents()
	if len(opponents) == 0 {
		return nil
	}
	switch n := rolls[0]; {
	case n >= 20:
		return ctx.PlayersSacrificeThen(opponents,
			effects.SacrificeSpec("a creature with the greatest power among creatures you control", effects.GreatestPowerYouControl()),
			"Myrkul's Edict — sacrifice a creature with the greatest power", myrkulsEdictDone, ctx.RunData(nil))
	case n >= 10:
		return ctx.PlayersSacrificeThen(opponents,
			effects.SacrificeSpec("a creature", effects.Creature()),
			"Myrkul's Edict — sacrifice a creature", myrkulsEdictDone, ctx.RunData(nil))
	}
	return effects.ChoosePlayer{
		Among:    effects.Opponents,
		Question: "Myrkul's Edict — choose an opponent; that player sacrifices a creature",
		ThenKey:  myrkulsEdictChosen,
	}.Apply(ctx)
})

// myrkulsEdictChosen is the 1—9 branch: the chosen opponent sacrifices a
// creature. An on-disk key.
var myrkulsEdictChosen = effects.ChoosePlayerThen("option-pick/myrkuls-edict-chosen", func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
	if chosen == uuid.Nil {
		return nil
	}
	return ctx.PlayersSacrificeThen([]uuid.UUID{chosen},
		effects.SacrificeSpec("a creature", effects.Creature()),
		"Myrkul's Edict — sacrifice a creature", myrkulsEdictDone, ctx.RunData(nil))
})
