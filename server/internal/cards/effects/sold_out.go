package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sold Out — Instant {3}{B}:
//
//	"Exile target creature. If it was dealt damage this turn, create a
//	 Clue token."
//
// "It was dealt damage this turn" is read off the creature as it is
// before the exile (DealtDamageThisTurn, the turn's damage events since
// it last entered), because once exiled it is a new object (CR 400.7).
// The Clue is the caster's whether or not a replacement kept the
// creature out of exile: the condition is about the damage, not the
// exile. It is not investigating (the ruling). An illegal target makes
// the spell do nothing at all (CR 608.2b).
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "6a69d4b1-72dd-4599-a16d-32299de5b0b9",
		Name:         "Sold Out",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			exile := ExileTarget{Target: target}
			if DealtDamageThisTurn()(ctx.Game, ctx.Controller(), c) {
				exile.ThenTail = ctx.ItemTail(soldOutClue)
			}
			return exile.Apply(ctx)
		},
	})
}

// soldOutClue is "create a Clue token", run once the exile has settled.
var soldOutClue = CardTail("sold-out/clue", func(ctx *Context, _ game.TailOutcome) error {
	return CreateToken{Template: ClueToken(), N: 1}.Apply(ctx)
})
