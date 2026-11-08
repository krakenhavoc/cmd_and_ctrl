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
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The cast trigger is Stormkeld Prowler's (WheneverYouCast
// with ManaValueGE, the mana value on the stack, CR 202.3e) with Fractal
// Mascot's tap-and-stun body (tapTargetAndStunIt); a target that has left
// by resolution is left alone (CR 608.2b). Storm Reading is a seven-mana
// spell, but the Giant is not on the battlefield while it is cast from
// hand, so it never triggers itself that way. Storm Reading is Careful
// Study's draw-then-discard: the discard prompt opens after the four
// draws, so the drawn cards are legal discards. The Adventure lifecycle
// is the engine's (CR 715.3d).
//
// No simplification.
const galvanicGiantOracleID = "48ab8ba0-0e85-4656-a798-ceb961fb02ce"

func init() {
	Register(Spec{
		OracleID:     galvanicGiantOracleID,
		Name:         "Galvanic Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(ManaValueGE(5), "Galvanic Giant — tap target creature an opponent controls and put a stun counter on it", tapTargetAndStunIt),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
	Register(Spec{
		OracleID:     galvanicGiantOracleID + "#1",
		Name:         "Storm Reading",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, Discards: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, item.Controller, item.SourceCardID, 4, 2, "")
		},
	})
}
