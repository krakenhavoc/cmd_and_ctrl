package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Minamo's Meddling — Instant {2}{U}{U} (Betrayers of Kamigawa, S82 OLD,
// #5602):
//
//	"Counter target spell. That spell's controller reveals their hand, then
//	 discards each card with the same name as a card spliced onto that
//	 spell."
//
// Counterspell's counter, then Cabal Therapy's reveal-and-discard. The
// spell's controller and the cards spliced onto it (StackItem.Spliced,
// CR 702.47a) are read before it is countered, since countering removes the
// item. The hand is revealed whether or not anything was spliced, and every
// card sharing a spliced card's name (CR 201.2) is discarded together, not
// by their choice; a spell with nothing spliced onto it discards nothing.
// Cards spliced onto it stay in hand (CR 702.47a), so they are among those
// discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c7239b0-e67f-4848-b244-8be9915bd2dc",
		Name:         "Minamo's Meddling",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			victim := uuid.Nil
			var names []string
			if v := ctx.Game.StackItemForEffect(stackID); v != nil {
				victim = v.Controller
				for _, s := range v.Spliced {
					names = append(names, s.Name)
				}
			}
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			p := ctx.Game.PlayerByIDForEffect(victim)
			if victim == uuid.Nil || p == nil {
				return nil
			}
			ctx.Game.RevealHandForEffect(victim)
			var named []uuid.UUID
			for _, c := range p.Hand.Cards {
				for _, n := range names {
					if game.CardNameMatches(c, n) {
						named = append(named, c.InstanceID)
						break
					}
				}
			}
			if len(named) == 0 {
				return nil
			}
			return ctx.Game.DiscardCardsKeyedForEffect(victim, named, ctx.Source(), game.KeyedTail{})
		},
	})
}
