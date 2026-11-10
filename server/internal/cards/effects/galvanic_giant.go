package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     galvanicGiantOracleID,
		Name:         "Galvanic Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventCast, YouCast(ManaValueGE(5)),
				"Galvanic Giant — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
	Register(Spec{
		OracleID:     galvanicGiantOracleID + "#1",
		Name:         "Storm Reading",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 4, 2, "")
		},
	})
}
