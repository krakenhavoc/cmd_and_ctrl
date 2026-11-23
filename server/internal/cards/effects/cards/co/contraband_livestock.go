package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Contraband Livestock — Instant {1}{W} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Exile target creature, then roll a d20.
//	 1—9 | Its controller creates a 4/4 green Ox creature token.
//	 10—19 | Its controller creates a 2/2 green Boar creature token.
//	 20 | Its controller creates a 0/1 white Goat creature token."
//
// The creature's controller is read before the exile (CR 608.2h) and rides
// the roll's tail as data; the d20 goes through the one dice door, and the
// token is made for that player, not the caster. A creature that was exiled
// by the spell or not (a replacement) still gets the roll, as printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbafe993-8118-40cc-9cb7-ee3f62a2c888",
		Name:         "Contraband Livestock",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			id := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.RollDice{Sides: 20, N: 1, Then: contrabandLivestockRolled, TailCarry: []uuid.UUID{controller}}.Apply(ctx)
		},
	})
}

// contrabandLivestockRolled reads the table: 1—9 an Ox, 10—19 a Boar, a 20 a
// Goat, for the exiled creature's controller. An on-disk key.
var contrabandLivestockRolled = effects.CardTail("dice/contraband-livestock-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Rolls) == 0 || len(o.Tail.Carry) == 0 {
		return nil
	}
	token := "4/4 green Ox"
	switch n := o.Rolls[0]; {
	case n >= 20:
		token = "0/1 white Goat"
	case n >= 10:
		token = "2/2 green Boar"
	}
	return effects.CreateToken{Controller: o.Tail.Carry[0], Template: effects.TokenCard(token), N: 1}.Apply(ctx)
})
