package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lay Down Arms — Sorcery for {W} (The Brothers' War):
//
//	"Exile target creature with mana value less than or equal to the number
//	of Plains you control. Its controller gains 3 life."
//
// The Plains are counted when the target is chosen and again at resolution
// (CR 608.2b): a Plains that left in response can make the target illegal.
func init() {
	Register(Spec{
		OracleID:     "aa91d092-7f4a-4e4d-85db-e200692f3d9d",
		Name:         "Lay Down Arms",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Completeness: CompletenessFull,
		Targets: TargetCreature("target creature with mana value less than or equal to the number of Plains you control",
			func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				return c.ManaValue() <= countControlled(g, caster, func(p game.Card) bool { return p.HasSubtypeWord("Plains") })
			}),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			controller := card.Controller
			if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), controller, 3)
		},
	})
}
