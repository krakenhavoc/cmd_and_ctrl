package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dramatic Reversal — Instant {1}{U} (EDHREC rank 660):
//
//	"Untap all nonland permanents you control."
//
// Half of the Isochron Scepter combo, and on its own a mana rock
// untapper at instant speed. Every nonland permanent the controller
// controls is untapped; lands stay as they are, which is the whole
// difference from a Turnabout. Untapping an already-untapped
// permanent is a no-op rather than an error.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da46904c-8fb8-44c2-b2ab-775a1cc12ec3",
		Name:         "Dramatic Reversal",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			me := ctx.Controller()
			var mine []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == me && !c.IsLand() && c.Tapped {
					mine = append(mine, c.InstanceID)
				}
			}
			for _, id := range mine {
				if err := (effects.UntapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
