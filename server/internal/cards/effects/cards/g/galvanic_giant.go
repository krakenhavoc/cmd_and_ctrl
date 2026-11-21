package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Giant // Storm Reading — Creature — Giant Wizard {3}{U}, 3/3 //
// Instant — Adventure {5}{U}{U} (Wilds of Eldraine, #4471):
//
//	Galvanic Giant — "Whenever you cast a spell with mana value 5 or
//	 greater, tap target creature an opponent controls and put a stun
//	 counter on it."
//	Storm Reading — "Draw four cards, then discard two cards. (Then exile
//	 this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). The cast trigger reads the spell's mana value on the stack, X
// counted (CR 202.3), and is Cryotheory Adept's tap and stun on a target
// chosen as the trigger goes on the stack. Storm Reading is Careful Study's
// draw-then-discard at four and two.
//
// No simplification.
const galvanicGiantOracleID = "48ab8ba0-0e85-4656-a798-ceb961fb02ce"

func init() {
	effects.Register(effects.Spec{
		OracleID:     galvanicGiantOracleID,
		Name:         "Galvanic Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, effects.YouCast(effects.ManaValueGE(5)),
				"Galvanic Giant — tap target creature an opponent controls and put a stun counter on it", effects.TapTargetAndStunIt),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     galvanicGiantOracleID + "#1",
		Name:         "Storm Reading",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 4, 2, "")
		},
	})
}
