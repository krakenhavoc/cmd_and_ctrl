package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tanuki Transplanter — Artifact Creature — Equipment Dog {3}{G}, 2/4:
//
//	"Whenever this creature or equipped creature attacks, add an amount
//	 of {G} equal to its power. Until end of turn, you don't lose this
//	 mana as steps and phases end.
//	 Reconfigure {3}"
//
// "Its power" is the attacking creature's, read as the trigger resolves
// (CR 608.2h): as it is then, or as it last existed if it has left. The
// mana carries Savage Ventmaw's keep mark (KeepManaUntilEndOfTurn).
// Reconfigure is reconfigure.go (#2639).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d1c7f42-19ef-405e-9edc-50a81b78f97c",
		Name:         "Tanuki Transplanter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisOrEquippedCreatureAttacks,
				"Tanuki Transplanter — add {G} equal to its power", tanukiTransplanterMana),
		},
		Activated: Reconfigure("{3}"),
	})
}

func tanukiTransplanterMana(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return AddMana{
		Produced: strings.Repeat("{G}", info.Power),
		Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
	}.Apply(ctx)
}
