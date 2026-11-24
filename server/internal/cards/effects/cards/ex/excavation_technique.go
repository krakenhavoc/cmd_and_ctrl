package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Excavation Technique — Sorcery {3}{W} (Commander 2021, #5504):
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do,
//	 choose an opponent to also copy it. Players may choose new targets
//	 for their copies.)
//	 Destroy target nonland permanent. Its controller creates two Treasure
//	 tokens."
//
// Demonstrate is the keyword's cast trigger (CR 702.144a, Creative
// Technique's). "Its controller" is read as the spell resolves, before
// the destroy, and they get the two Treasures whether or not the
// permanent was destroyed (indestructible, regeneration). A target gone by
// resolution makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bcd3e6a1-7b76-408d-9236-3791c3e1064a",
		Name:         "Excavation Technique",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Demonstrate()},
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Controller: controller, Template: effects.TreasureToken(), N: 2}.Apply(ctx)
		},
	})
}
