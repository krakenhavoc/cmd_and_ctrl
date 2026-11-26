package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Reckoning — Instant {B} (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Exile target colorless creature. You gain life equal to its power."
//
// Banish from Edoras's exile of the one target. The life is its power as it
// last existed on the battlefield (CR 608.2h), read just before the exile so
// an anthem or a counter is counted. A target that is no longer a colorless
// creature on resolution is illegal and the spell does nothing (CR 608.2b); a
// creature with no or negative power gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32aa920c-0d91-4bc9-803c-d13857688d10",
		Name:         "Infernal Reckoning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target colorless creature", effects.Colorless()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			power := 0
			if c, found := ctx.Game.LookupCardForEffect(id); found {
				power = c.CurrentPower()
			}
			if power <= 0 {
				return effects.ExileTarget{Target: id}.Apply(ctx)
			}
			return effects.ExileTarget{Target: id, ThenTail: ctx.TailN(infernalReckoningLife, power)}.Apply(ctx)
		},
	})
}

// infernalReckoningLife is "You gain life equal to its power", told that the
// creature really reached exile: a move the CR 614 window cancelled or
// redirected gains nothing. The power rides the tail as N, read before the
// exile. An on-disk key.
var infernalReckoningLife = effects.CardTail("infernal-reckoning/gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: o.Tail.Player, Amount: o.Tail.N}.Apply(ctx)
})
