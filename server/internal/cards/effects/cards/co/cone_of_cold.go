package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cone of Cold — Sorcery {3}{U}:
//
//	"Roll a d20.
//	 1—9 | Tap all creatures your opponents control.
//	 10—19 | Tap all creatures your opponents control. Those creatures
//	 don't untap during their controllers' next untap steps.
//	 20 | Tap all creatures your opponents control. Those creatures don't
//	 untap during their controllers' next untap steps. Until your next
//	 turn, creatures your opponents control enter tapped."
//
// The roll is the shared d20 (CR 706); a result modifier applies before
// the row is read. "Those creatures" are the creatures the opponents
// control as the spell taps them (CR 608.2c), each marked to stay tapped
// through its controller's next untap step (TapAndFreeze), whether or not
// it was already tapped.
//
// The 20 row's last sentence is the #6093 timed replacement
// (game.ModCreaturesEnterTappedForOpponents): until your next turn, every
// creature entering under an opponent's control enters tapped (CR
// 614.1c), a token or a card, cast or put onto the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56fd0752-9fb0-4c05-90e5-c3b268b57c6f",
		Name:         "Cone of Cold",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RollDice{Sides: 20, N: 1, Then: coneOfColdRolled}.Apply(ctx)
		},
	})
}

const coneOfColdFreezeLabel = "Cone of Cold — those creatures don't untap during their controllers' next untap steps"

var coneOfColdRolled = effects.DiceTail("dice/cone-of-cold-rolled", func(ctx *effects.Context, rolls []int) error {
	n := rolls[0]
	if n < 10 {
		return effects.TapAllCreaturesYourOpponentsControl(ctx.Game, ctx.Item)
	}
	var ids []uuid.UUID
	for _, opp := range ctx.Opponents() {
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.Controller == opp && c.IsCreature() {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if err := (effects.TapAndFreeze{Targets: ids, Label: coneOfColdFreezeLabel}).Apply(ctx); err != nil {
		return err
	}
	if n >= 20 {
		ctx.Game.CreaturesEnterTappedForOpponentsForEffect(ctx.Source(), ctx.Controller(),
			effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
			"Cone of Cold — until your next turn, creatures your opponents control enter tapped")
	}
	return nil
})
