package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Wrath of Khan — Enchantment — Saga {2}{R} (Star Trek, #4973):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — This Saga deals 4 damage to target creature an opponent
//	     controls. At the beginning of the next combat this turn, gain
//	     control of that creature until end of turn. Untap it. It gains
//	     haste until end of turn.
//	 II — You may exile a card from your hand. If you do, draw a card.
//	 III — Each opponent sacrifices a nontoken permanent of their
//	     choice."
//
// I deals the damage, then creates the delayed trigger (CR 603.7a). The
// trigger is event-conditioned on the next EventStepBegan of a
// beginning-of-combat step: it fires once (CR 603.7b), and the "this
// turn" duration ScheduleDelayedTriggerForEffect stamps on an
// event-conditioned trigger drops it at cleanup if no combat began, so
// a chapter resolved after combat does nothing. It remembers the
// creature as an object (ID and CR 400.7 epoch): a creature that
// survived the 4 damage and is still that object on the battlefield is
// taken, untapped and given haste until end of turn (threatenAndGrant);
// one that died, or left and came back, is a new object and is not
// (CR 603.7c). A chapter whose target became illegal does not resolve
// (CR 608.2b), so no delayed trigger is made.
//
// II asks for up to one card of your hand; one that reached exile draws
// a card, and none drawn otherwise. III is the each-opponent edict
// (EachPlayerSacrifices) over nontoken permanents, each opponent
// choosing their own (CR 701.21a).
//
// No simplification.
func init() {
	const name = "The Wrath of Khan"
	Register(Spec{
		OracleID:     "22249955-0741-470e-bc2b-f32d4510849b",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "4 damage to target creature an opponent controls; at the beginning of the next combat this turn, gain control of it"),
				TargetCreature("target creature an opponent controls", OpponentControls()), theWrathOfKhanChapterOne),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "you may exile a card from your hand; if you do, draw a card"),
				theWrathOfKhanChapterTwo),
			TriggerEdict(ChapterTrigger(3, SagaChapterLabel(name, 3, "each opponent sacrifices a nontoken permanent"),
				Do(EachPlayerSacrifices{
					ExceptController: true,
					Match:            Not(IsTokenPredicate()),
					Label:            "a nontoken permanent",
				}))),
		},
	})
}

// theWrathOfKhanChapterOne is chapter I: the damage, then the delayed
// trigger naming the creature as the object it was.
func theWrathOfKhanChapterOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 || legal[0].Kind != game.TargetCard {
		return nil
	}
	target := legal[0].ID
	ref, ok := g.PermanentRefForEffect(target)
	if !ok {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: target, Amount: 4}).Apply(ctx); err != nil {
		return err
	}
	return DelayedOnEvent{
		Label:     "The Wrath of Khan — gain control of that creature until end of turn, untap it, it gains haste",
		On:        []game.EventKind{game.EventStepBegan},
		Condition: theWrathOfKhanNextCombatCondition,
		Cards:     []uuid.UUID{target},
		Body:      theWrathOfKhanTakeThatCreatureBody,
		Params:    game.EffectParams{Objects: []game.ObjectRef{ref}},
	}.Apply(ctx)
}

var (
	// "At the beginning of the next combat this turn": a beginning of
	// combat step began. The trigger fires on the first one (CR 603.7b)
	// and expires with the turn. An on-disk key.
	theWrathOfKhanNextCombatCondition = game.DelayedCondition("combat/the-wrath-of-khan-next-beginning-of-combat",
		func(ev game.Event, _ *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventStepBegan && ev.Step == game.StepBeginCombat
		})

	// "Gain control of that creature until end of turn. Untap it. It
	// gains haste until end of turn." — only while the named object is
	// still that object on the battlefield (CR 400.7, 603.7c). An
	// on-disk key.
	theWrathOfKhanTakeThatCreatureBody = game.DelayedBody("the-wrath-of-khan/gain-control-of-that-creature",
		func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
			for _, ref := range p.Objects {
				c, ok := battlefieldCardForEffect(g, ref.ID)
				if !ok || c.ObjectEpoch != ref.Epoch {
					continue
				}
				if err := threatenAndGrant(NewContext(g, item), ref.ID, "The Wrath of Khan", "haste"); err != nil {
					return err
				}
			}
			return nil
		})
)

// theWrathOfKhanChapterTwo is chapter II: up to one card of the
// controller's hand. An empty hand has nothing to exile, so nothing is
// drawn.
func theWrathOfKhanChapterTwo(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	var hand []uuid.UUID
	if p := g.PlayerByIDForEffect(player); p != nil && p.Hand != nil {
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
	}
	if len(hand) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "The Wrath of Khan — you may exile a card from your hand. If you do, draw a card.",
		Cards:    hand,
		Min:      0,
		Max:      1,
		ThenKey:  theWrathOfKhanExileFromHand,
		LiveZone: game.ZoneHand,
	})
	return nil
}

var (
	// The keyed answer to chapter II's "you may exile a card from your
	// hand". An on-disk key.
	theWrathOfKhanExileFromHand = game.RegisterChooseCardsThen("choose-cards/the-wrath-of-khan-exile-from-hand",
		func(g *game.Game, r game.CardsChosen) error {
			if len(r.Picked) == 0 {
				return nil
			}
			return g.ExileCardKeyedForEffect(r.Picked[0],
				game.KeyedTail{Ref: theWrathOfKhanDrawIfExiled, Player: r.Chooser, Source: r.Source})
		})

	// "If you do, draw a card": only when the card reached exile. An
	// on-disk key.
	theWrathOfKhanDrawIfExiled = game.RegisterTail("the-wrath-of-khan/draw-if-exiled",
		func(g *game.Game, o game.TailOutcome) error {
			if !oneWentThisWay(o) {
				return nil
			}
			return g.DrawNForEffect(o.Tail.Player, 1)
		})
)
