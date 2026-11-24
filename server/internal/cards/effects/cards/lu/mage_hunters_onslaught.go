package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mage Hunters' Onslaught — Sorcery {2}{B}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"Destroy target creature or planeswalker.
//	 Whenever a creature blocks this turn, its controller loses 1 life."
//
// The destroy is Baleful Mastery's target step. The second sentence is a
// delayed triggered ability the spell creates (CR 603.7a) as it resolves, one
// that watches every blocker of every player until end of turn (CR 603.7b,
// Benefactor's Draught's EventBlock shape): it triggers once per blocking
// creature however many attackers it blocks (CR 509.3a), and its controller is
// the spell's caster while the life is lost by the blocking creature's
// controller (read from the blocker, which rides the fired item as its payload). If the target is illegal the spell does not resolve, so the
// trigger is not created either (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22195e0e-a426-48d6-b100-480c7e2bd5cc",
		Name:         "Mage Hunters' Onslaught",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			if err := effects.DestroyFirstLegalTarget(ctx.Game, ctx.Item); err != nil {
				return err
			}
			return effects.DelayedOnEvent{
				Label:     "Mage Hunters' Onslaught — its controller loses 1 life",
				On:        []game.EventKind{game.EventBlock},
				Condition: oldB50AnyCreatureBlocks,
				Body:      oldB50BlockerControllerLosesLife,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}

var oldB50AnyCreatureBlocks = game.DelayedCondition("old-b050/any-creature-blocks", func(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventBlock && ev.Amount <= 1
})

// oldB50BlockerControllerLosesLife is the fired trigger's body: the blocker
// rides the item as its payload (a delayed trigger's item carries no event),
// and its controller loses 1 life. A blocker that has left the game is read as
// the card last was.
var oldB50BlockerControllerLosesLife = game.SimpleDelayedBody("old-b050/blocker-controller-loses-1", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	blockers := ctx.PayloadCards()
	if len(blockers) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(blockers[0])
	if !ok || c.Controller == uuid.Nil {
		return nil
	}
	if p := g.PlayerByIDForEffect(c.Controller); p == nil || p.Eliminated {
		return nil
	}
	return g.ChangePlayerLifeForEffect(item.SourceCardID, c.Controller, -1)
})
