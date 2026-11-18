package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mr. Orfeo, the Boulder — Legendary Creature — Rhino Warrior {1}{B}{R}{G},
// 2/4 (Streets of New Capenna, S82):
//
//	"Whenever you attack, double target creature's power until end of turn."
//
// "Whenever you attack" is Araña's shape: once per declaration
// (OncePerBatch), the target chosen as the trigger goes on the stack. The
// target is any creature, not only an attacker. The doubling reads the
// power as the trigger resolves (CR 701.10b, DoublePowerUntilEOT) and does
// nothing if the target is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58447f16-1df9-4e9a-bb22-b9aad042055a",
		Name:         "Mr. Orfeo, the Boulder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Mr. Orfeo, the Boulder — double target creature's power until end of turn",
					Do(Instruction(func(ctx *Context) error {
						id := FirstLegalBattlefieldTarget(ctx)
						if id == uuid.Nil {
							return nil
						}
						return DoublePowerUntilEOT(ctx, id, "Mr. Orfeo, the Boulder — double power")
					}))),
				TargetCreature("target creature"))),
		},
	})
}
