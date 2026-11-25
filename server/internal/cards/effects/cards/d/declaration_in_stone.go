package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Declaration in Stone — Sorcery {1}{W}:
//
//	"Exile target creature and all other creatures its controller
//	 controls with the same name as that creature. That player
//	 investigates for each nontoken creature exiled this way."
//
// The target's controller and its name are read as the spell resolves
// (last-known information if the creature has already left, CR 608.2h);
// the sweep is one simultaneous exile of the target and of every other
// creature that controller controls with that name (CR 201.2). A
// face-down creature has no name (CR 708.2a), so it matches nothing but
// itself. The Clues go to that player, one for each nontoken creature
// that really reached exile, counted from the batch's pre-move copies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49554441-bb48-4829-b8d6-55e687a6c2e2",
		Name:         "Declaration in Stone",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, err := effects.OldB157ExileTargetAndSameName(ctx, ctx.Tail(declarationInStoneClues))
			return err
		},
	})
}

// declarationInStoneClues is "that player investigates for each nontoken
// creature exiled this way": each swept copy carries the controller it
// had on the battlefield.
var declarationInStoneClues = effects.MassTail("mass/declaration-in-stone-clues", func(ctx *effects.Context, swept []game.Card, _ int) error {
	counts := map[uuid.UUID]int{}
	var order []uuid.UUID
	for _, c := range swept {
		if c.IsToken() {
			continue
		}
		if counts[c.Controller] == 0 {
			order = append(order, c.Controller)
		}
		counts[c.Controller]++
	}
	for _, p := range order {
		if err := (effects.Investigate{Player: p, N: counts[p]}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
