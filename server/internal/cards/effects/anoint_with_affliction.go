package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anoint with Affliction — Instant {1}{B}:
//
//	"Exile target creature if it has mana value 3 or less.
//	 Corrupted — Exile that creature instead if its controller has three or
//	 more poison counters."
//
// Both are read as the spell resolves: the creature's mana value, and the
// poison of the player who controls it then. Corrupted widens the clause
// ("instead"), so a poisoned player's creature of any mana value is
// exiled. This is the creature's controller's poison, not the caster's
// opponents'. A target that became illegal fizzles the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "402c3afe-7041-41a3-8fcd-64a060c12d4a",
		Name:         "Anoint with Affliction",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			mv, _ := ctx.Game.ManaValueForEffect(c)
			if mv > 3 && !playerIsCorrupted(ctx.Game, c.Controller) {
				return nil
			}
			return ExileTarget{Target: id}.Apply(ctx)
		},
	})
}
