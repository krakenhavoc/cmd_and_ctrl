package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sound the Trumpets — Instant {1}{U}{U}:
//
//	"Counter target spell. If that spell's mana value was 2 or less,
//	 recruit. (Draw a card, then discard a card. If you discarded a
//	 nonland card, create a 1/1 white Human Soldier creature token.)"
//
// "That spell's mana value" is read off the spell on the stack before
// the counter moves it, with {X} counted as announced (CR 202.3e), and
// filed for the second step, which runs only once the counter has
// settled (a countered commander may ask where it goes, CR 903.9). The
// recruit hangs on the mana value, not on the counter succeeding, so a
// spell that can't be countered still lets you recruit; a target that is
// gone by resolution means the whole spell does nothing (CR 608.2b).
//
// No simplification. The Hobbit (#4475).
const soundTheTrumpetsRecord = "sound-the-trumpets/mana-value"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "dbcf2a79-b455-43ac-9cf0-9e63708800f4",
		Name:         "Sound the Trumpets",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					ctx.RecordStep(soundTheTrumpetsRecord, nil, effects.SpellManaValueForEffect(ctx.Game, t.ID))
					return effects.CounterTarget{StackID: t.ID}.Apply(ctx)
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				r := ctx.StepRecord(soundTheTrumpetsRecord)
				if len(r.Ns) == 0 || r.N() > 2 {
					return nil
				}
				return effects.Recruit{}.Apply(ctx)
			}),
		),
	})
}
