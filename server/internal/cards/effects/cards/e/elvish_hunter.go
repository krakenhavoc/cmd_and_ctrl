package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elvish Hunter — Creature — Elf Archer {1}{G}, 1/1 (Fallen Empires, S82 OLD batch
// 852, #5637):
//
//	"{1}{G}, {T}: Target creature doesn't untap during its controller's next untap
//	 step."
//
// House Guildmage's freeze on a creature chosen as the ability goes on the stack:
// the clause is pinned to the object it names as it resolves (a creature that
// has left is a new object, CR 400.7) and skips that creature's own controller's
// next untap step (CR 502.3). A target that is illegal by resolution does
// nothing (CR 608.2b). The {T} is a cost, so a summoning-sick Hunter can't pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49b351d8-8b63-4f9e-a323-344ed35fdfff",
		Name:         "Elvish Hunter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}, {T}: Target creature doesn't untap during its controller's next untap step.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.DoesntUntapNextUntapStep{Targets: []uuid.UUID{id}, Label: "Elvish Hunter — doesn't untap during its controller's next untap step"}.Apply(ctx)
			},
		}},
	})
}
