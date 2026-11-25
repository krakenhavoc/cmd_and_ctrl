package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Promise of Tomorrow — Enchantment {2}{W} (Commander Legends, #5507):
//
//	"Whenever a creature you control dies, exile it.
//	 At the beginning of each end step, if you control no creatures,
//	 sacrifice this enchantment and return all cards exiled with it to the
//	 battlefield under your control."
//
// The dies trigger exiles the card from the graveyard it went to, linked to
// this object (CR 607.2a, ExileTarget.Linked); a token is gone from the
// graveyard (CR 111.7) and a card that something has already moved out of
// the graveyard is left alone. The end step trigger is an intervening "if"
// (CR 603.4, CondBoard: asked as the step begins and as it resolves). It
// sacrifices the enchantment, if it is still this object on the
// battlefield, and then returns every card still in exile linked to it
// (last-known information, CR 608.2h: the link is read after the
// sacrifice, as Savior of Ollenbock's is) as one entry, all of them under
// the controller's control, not their owners' (CR 110.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26c22894-10e5-48f1-a32a-a7f1f3f7e12e",
		Name:         "Promise of Tomorrow",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.ACreatureYouControlDied,
				"Promise of Tomorrow — exile it", promiseOfTomorrowExile),
			effects.InterveningIf(effects.CondBoard{
				Name: "you control no creatures",
				Test: promiseOfTomorrowNoCreatures,
			}, effects.On(game.EventBeginEndStep, effects.AnyPlayer,
				"Promise of Tomorrow — sacrifice it and return all cards exiled with it to the battlefield under your control",
				promiseOfTomorrowReturn)),
		},
	})
}

// promiseOfTomorrowExile exiles the creature that died, linked to this
// object, if it is still in a graveyard.
func promiseOfTomorrowExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id := ctx.Trigger().Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ExileTarget{Target: id, Linked: true}.Apply(ctx)
}

// promiseOfTomorrowNoCreatures is the intervening "if you control no
// creatures".
func promiseOfTomorrowNoCreatures(ctx *effects.Context) bool {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() {
			return false
		}
	}
	return true
}

// promiseOfTomorrowReturn sacrifices the enchantment if it is still this
// object, then returns what is linked to it, whether or not it was there to
// sacrifice (both instructions are the controller's to carry out).
func promiseOfTomorrowReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if effects.B09SourceStillOnBattlefield(g, item) {
		if err := g.SacrificePermanentForEffect(item.SourceCardID); err != nil {
			return err
		}
	}
	return promiseOfTomorrowReturnLinked(ctx)
}

func promiseOfTomorrowReturnLinked(ctx *effects.Context) error {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	return effects.ReturnFromExileTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
}
