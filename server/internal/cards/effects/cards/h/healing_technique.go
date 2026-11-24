package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Healing Technique — Sorcery {3}{G} (Commander 2021, #5504):
//
//	"Demonstrate (When you cast this spell, you may copy it. If you do,
//	 choose an opponent to also copy it. Players may choose new targets
//	 for their copies.)
//	 Return target card from your graveyard to your hand. You gain life
//	 equal to that card's mana value. Exile Healing Technique."
//
// Demonstrate is the keyword's cast trigger (CR 702.144a, first used by
// Creative Technique). Each copy resolves for its own controller, so the
// opponent who copied it returns a card from THEIR graveyard and gains the
// life. The life is the card's mana value as it sat in the graveyard (CR
// 608.2h, an X is 0, a land is 0). The last sentence exiles the spell: a
// copy is not a card and simply ceases to exist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b99c32d-1d02-410e-84bf-9db2ca05f576",
		Name:         "Healing Technique",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Demonstrate()},
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target card from your graveyard", effects.YouOwn()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			mv := card.ManaValue()
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GainLife{Player: ctx.Controller(), Amount: mv}).Apply(ctx); err != nil {
				return err
			}
			return effects.ExileTarget{Target: ctx.Source()}.Apply(ctx)
		},
	})
}
