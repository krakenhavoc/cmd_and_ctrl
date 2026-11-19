package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overpowering Attack — Sorcery {3}{R}{R}:
//
//	"Freerunning {2}{R} (You may cast this spell for its freerunning
//	 cost if you dealt combat damage to a player this turn with an
//	 Assassin or commander.)
//	 Untap all creatures you control that attacked this turn. If it's
//	 your main phase, there is an additional combat phase after this
//	 phase, followed by an additional main phase."
//
// The untap reads the per-turn attack history per object (a creature that
// left and came back has not attacked, CR 400.7) and only the creatures its
// controller controls. The extra combat and main are Aggravated Assault's
// (CR 500.8), added only when the spell resolves in its controller's own
// main phase; freerunning is Eagle Vision's alternative cost.
func init() {
	Register(Spec{
		OracleID:         "e31235f0-1f25-4158-b46e-cacde2f58b7e",
		Name:             "Overpowering Attack",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{ExtraCombat: 1},
		AlternativeCosts: []game.AlternativeCost{Freerunning("{2}{R}", "Assassin")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			you := ctx.Controller()
			var mine []uuid.UUID
			for _, id := range CreaturesThatAttackedThisTurn(ctx.Game) {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.Controller == you {
					mine = append(mine, id)
				}
			}
			if err := untapEach(ctx, mine); err != nil {
				return err
			}
			if !b14InYourMainPhase(ctx.Game, you) {
				return nil
			}
			return ExtraCombatAndMainAfterThisMain().Apply(ctx)
		},
	})
}
