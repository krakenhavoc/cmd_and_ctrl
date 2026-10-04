package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Diplomacy of the Wastes — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card. If you control a Warrior, that
//	 player loses 2 life."
//
// The revealed-hand pick (ADR 0116) filtered to nonland cards, then the
// Warrior check on the next line (ADR 0116 §6). The check reads your
// battlefield, not the pick, so it cannot depend on which card is
// chosen or whether one is: the 2014-11-24 ruling says the opponent
// loses 2 life with a Warrior even when nothing was discarded
// (CR 609.3). "That player" is the target, which is still legal here
// or the spell would not have resolved (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b4670b9-6701-4ca6-afd7-256742267e86",
		Name:         "Diplomacy of the Wastes",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			victim := TargetedPlayer(ctx)
			if err := (ChooseFromRevealedHand{
				Player: victim,
				Filter: Nonland(),
				Label:  "nonland card",
			}).Apply(ctx); err != nil {
				return err
			}
			if victim == uuid.Nil || !ControlsA("Warrior")(ctx.Game, ctx.Controller()) {
				return nil
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -2)
		},
	})
}
