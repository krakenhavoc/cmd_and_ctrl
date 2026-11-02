package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Penance Stare — Instant {W}:
//
//	"Destroy target attacking creature with mana value 3 or less."
//
// Doom Blade's shape with a combat-state and a mana-value filter on the
// target. Both are read from the creature's effective characteristics,
// when the spell is announced and again as it resolves (CR 608.2b), so a
// creature that stops attacking in response is not destroyed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e21a1ad5-bf31-425f-92c2-7818bdb1cef8",
		Name:         "Penance Stare",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetCreature("target attacking creature with mana value 3 or less",
			And(AttackingCreature(), ManaValueLE(3))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
