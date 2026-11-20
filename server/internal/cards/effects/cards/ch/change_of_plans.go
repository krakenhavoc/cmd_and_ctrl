package cardsch

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Change of Plans — Instant {X}{1}{U} (Streets of New Capenna, #5491):
//
//	"Each of X target creatures you control connive. You may have any number of
//	 them phase out. (To have a creature connive, draw a card, then discard a
//	 card. If you discarded a nonland card, put a +1/+1 counter on that
//	 creature. Treat phased-out permanents and anything attached to them as
//	 though they don't exist until their controller's next turn.)"
//
// Exactly X targets (CountFromX), each still legal as it resolves (CR 608.2b).
// The targets connive one after another in the order they were chosen, each
// finishing its draw, discard and counter before the next begins (CR 701.50a,
// connive.go's chain; each connive opens its CR 614 window, so Leader,
// Super-Genius still replaces it). Then, once the last has connived, you choose
// any number of those creatures still on the battlefield, none included, and
// they phase out together with what is attached to them (CR 702.26).
//
// The chain is Carry-only data: the creature conniving, the ones still to go,
// a uuid.Nil, then every target, so a table waiting on a discard or on the
// phase-out choice is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2f43df5-ec64-4375-b67f-b198732ab0b1",
		Name:         "Change of Plans",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetsCountedByX(effects.TargetCreature("X target creatures you control", effects.YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return changeOfPlansNext(ctx.Game, ctx.Controller(), ids, ids)
		},
	})
	changeOfPlansTail = game.RegisterTail("connive/change-of-plans", func(g *game.Game, o game.TailOutcome) error {
		cur, rest, all, ok := changeOfPlansSplit(o.Tail.Carry)
		if !ok {
			return nil
		}
		if o.Tail.N <= 0 { // the connive window cancelled it
			return changeOfPlansNext(g, o.Tail.Player, rest, all)
		}
		return effects.ConniveDrawAndDiscard(g, o.Tail.Player, cur, o.Amount, o.Tail.N,
			"Connive: discard a card", changeOfPlansDiscarded, changeOfPlansJoin(rest, all))
	})
	changeOfPlansDiscarded = game.RegisterDiscardThen("choose-cards/change-of-plans-discarded", func(g *game.Game, r game.DiscardsDone) error {
		creature := r.Source
		if discarded := r.Discarded.By(r.Player); len(discarded) > 0 {
			if card, ok := g.LookupCardForEffect(discarded[0]); ok && !card.IsLand() && effects.OnBattlefield(g, creature) {
				if err := g.AddCounterForEffect(creature, game.CounterPlusOne, 1); err != nil {
					return err
				}
			}
		}
		// CR 701.50f: connived once the counter step is done.
		g.ConnivedForEffect(creature, r.Player)
		rest, all, ok := changeOfPlansSplitRest(r.Carry)
		if !ok {
			return nil
		}
		return changeOfPlansNext(g, r.Player, rest, all)
	})
	changeOfPlansPhaseOut = game.RegisterChooseCardsThen("choose-cards/change-of-plans-phase-out", func(g *game.Game, r game.CardsChosen) error {
		var out []uuid.UUID
		for _, id := range r.Picked {
			if effects.OnBattlefield(g, id) {
				out = append(out, id)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return g.PhaseOutForEffect(r.Source, out...)
	})
}

// changeOfPlansTail is the connive window's tail, changeOfPlansDiscarded the
// discard's, and changeOfPlansPhaseOut the phase-out pick's. Each is registered
// in init, because the first two go on to the next creature and so name
// themselves. The keys are on-disk identities: never renamed, never reused.
var (
	changeOfPlansTail      game.TailRef
	changeOfPlansDiscarded game.DiscardThen
	changeOfPlansPhaseOut  game.ChooseCardsThen
)

// changeOfPlansJoin is the chain's Carry: the creatures still to go, a
// uuid.Nil, then every target.
func changeOfPlansJoin(rest, all []uuid.UUID) []uuid.UUID {
	out := append([]uuid.UUID(nil), rest...)
	out = append(out, uuid.Nil)
	return append(out, all...)
}

// changeOfPlansSplitRest undoes changeOfPlansJoin.
func changeOfPlansSplitRest(carry []uuid.UUID) (rest, all []uuid.UUID, ok bool) {
	i := slices.Index(carry, uuid.Nil)
	if i < 0 {
		return nil, nil, false
	}
	return carry[:i:i], carry[i+1:], true
}

// changeOfPlansSplit undoes a tail's Carry: the creature conniving, the ones
// still to go, and every target.
func changeOfPlansSplit(carry []uuid.UUID) (cur uuid.UUID, rest, all []uuid.UUID, ok bool) {
	if len(carry) == 0 || carry[0] == uuid.Nil {
		return uuid.Nil, nil, nil, false
	}
	rest, all, ok = changeOfPlansSplitRest(carry[1:])
	return carry[0], rest, all, ok
}

// changeOfPlansNext makes the first of ids connive, or, with none left, asks
// which of the targets still on the battlefield phase out.
func changeOfPlansNext(g *game.Game, controller uuid.UUID, ids, all []uuid.UUID) error {
	if len(ids) == 0 {
		return changeOfPlansAskPhaseOut(g, controller, all)
	}
	carry := append([]uuid.UUID{ids[0]}, changeOfPlansJoin(ids[1:], all)...)
	return g.ConniveForEffect(controller, ids[0], 1, game.KeyedTail{
		Ref:    changeOfPlansTail,
		Player: controller,
		Carry:  carry,
	})
}

// changeOfPlansAskPhaseOut is "you may have any number of them phase out".
func changeOfPlansAskPhaseOut(g *game.Game, controller uuid.UUID, all []uuid.UUID) error {
	var cands []uuid.UUID
	for _, id := range all {
		if effects.OnBattlefield(g, id) {
			cands = append(cands, id)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   cands[0],
		Question: "Change of Plans — choose any number of those creatures to phase out",
		Cards:    cands,
		Min:      0,
		Max:      len(cands),
		Zone:     game.ZoneBattlefield,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  changeOfPlansPhaseOut,
	})
	return nil
}
