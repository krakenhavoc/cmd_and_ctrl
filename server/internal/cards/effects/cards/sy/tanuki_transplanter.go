package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "4d1c7f42-19ef-405e-9edc-50a81b78f97c",
		Name:         "Tanuki Transplanter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisOrEquippedCreatureAttacks,
				"Tanuki Transplanter — add {G} equal to its power", tanukiTransplanterMana),
		},
		Activated: effects.Reconfigure("{3}"),
	})
}

func tanukiTransplanterMana(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Power <= 0 {
		return nil
	}
	return effects.AddMana{
		Amount: effects.ManaN(info.Power, "G"),
		Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()},
	}.Apply(ctx)
}
