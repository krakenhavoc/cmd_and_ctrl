package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summary Dismissal — Instant {2}{U}{U} (Innistrad Remastered):
//
//	"Exile all other spells and counter all abilities."
//
// Every other spell on the stack is exiled (not countered: a spell that can't
// be countered is still exiled, and nothing watching "whenever a spell is
// countered" fires), and every activated or triggered ability on the stack is
// countered (CR 701.6). An ability that can't be countered stays. The items
// are read once, as the spell resolves; things put on the stack afterwards
// are not affected.
func init() {
	Register(Spec{
		OracleID:     "f69d65e0-4d8a-4e27-b4dd-dd8c50acb8eb",
		Name:         "Summary Dismissal",
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var spells, abilities []uuid.UUID
			for id, other := range g.StackMeta {
				if other == nil || id == item.ID {
					continue
				}
				switch other.Kind {
				case game.StackItemSpell:
					spells = append(spells, id)
				case game.StackItemActivated, game.StackItemTriggered:
					abilities = append(abilities, id)
				}
			}
			sort.Slice(spells, func(i, j int) bool { return spells[i].String() < spells[j].String() })
			sort.Slice(abilities, func(i, j int) bool { return abilities[i].String() < abilities[j].String() })
			for _, id := range spells {
				if err := g.ExileSpellForEffect(id); err != nil {
					return err
				}
			}
			for _, id := range abilities {
				if err := (CounterTarget{StackID: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
