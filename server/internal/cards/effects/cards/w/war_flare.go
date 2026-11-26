package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War Flare — {2}{R}{W} Instant (Fate Reforged):
//
//	"Creatures you control get +2/+1 until end of turn. Untap those
//	 creatures."
//
// The set is the creatures you control as the spell resolves (CR 611.2c); the
// same creatures are untapped.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7146eecb-8a79-467c-936e-f577e481cf59",
		Name:         "War Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller != ctx.Controller() || !c.IsCreature() {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: c.InstanceID, Power: 2, Toughness: 1, Label: "War Flare — +2/+1"}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
				if err := (effects.UntapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
