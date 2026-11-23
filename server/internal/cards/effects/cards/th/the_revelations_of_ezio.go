package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Revelations of Ezio — Enchantment — Saga {2}{B} (Assassin's Creed,
// #5948):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Destroy target tapped creature an opponent controls.
//	 II — Whenever an Assassin you control attacks this turn, put a +1/+1
//	 counter on it.
//	 III — Return target Assassin creature card from your graveyard to the
//	 battlefield with an additional +1/+1 counter on it."
//
// Chapter I is a destroy over a tapped creature an opponent controls, chosen as
// the chapter goes on the stack. Chapter II creates a delayed triggered ability
// (CR 603.7a) that watches Assassins you control being declared as attackers until
// the turn ends (First Day of Class's shape), once for each, carrying the attacker as
// its payload; the counter goes on only if it is still the same object on the
// battlefield as the trigger resolves (CR 603.7c). An Assassin that was already
// attacking when the chapter resolved is not affected: the ability did not exist
// when it attacked. Chapter III returns the target with one +1/+1 counter as it
// enters (a CR 614.1c "enters with", so a counter replacement sees it).
//
// No simplification.
func init() {
	const name = "The Revelations of Ezio"
	effects.Register(effects.Spec{
		OracleID:     "c7fe1f05-f412-4991-98bd-f6a4b8b62501",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTriggerTargeting(1, effects.SagaChapterLabel(name, 1, "destroy target tapped creature an opponent controls"),
				effects.TargetCreature("target tapped creature an opponent controls", effects.OpponentControls(), effects.B751Tapped()),
				effects.DestroyFirstLegalTarget),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "whenever an Assassin you control attacks this turn, put a +1/+1 counter on it"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.DelayedOnEvent{
						Label:     name + " — put a +1/+1 counter on the attacking Assassin",
						On:        []game.EventKind{game.EventAttack},
						Condition: revelationsOfEzioAttackCondition,
						Body:      revelationsOfEzioCounterBody,
						Repeats:   true,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target Assassin creature card from your graveyard to the battlefield with an additional +1/+1 counter on it"),
				effects.TargetCardInGraveyard("target Assassin creature card from your graveyard", effects.YouOwn(), effects.Creature(), effects.HasSubtype("Assassin")),
				revelationsOfEzioReturn),
		},
	})
}

var (
	// "Whenever an Assassin you control attacks": the creature an EventAttack
	// names, while it is an Assassin the delayed trigger's controller controls.
	// An on-disk key.
	revelationsOfEzioAttackCondition = game.DelayedEachCondition("attack/each-assassin-you-control",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) []uuid.UUID {
			if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
				return nil
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || c.Controller != dt.Controller || !effects.HasSubtype("Assassin")(g, dt.Controller, c) {
				return nil
			}
			return []uuid.UUID{ev.CardID}
		})

	// "Put a +1/+1 counter on it": the payload's permanent, if it is still that
	// object on the battlefield (CR 603.7c). An on-disk key.
	revelationsOfEzioCounterBody = game.SimpleDelayedBody("counter/plus-one-on-the-attacking-assassin", func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		obj, ok := effects.PayloadObject(ctx)
		if !ok {
			return nil
		}
		c, ok := effects.BattlefieldCardForEffect(g, obj.ID)
		if !ok || c.ObjectEpoch != obj.Epoch {
			return nil
		}
		return effects.AddCounter{Target: obj.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	})
)

// revelationsOfEzioReturn is chapter III: the target, if still legal, returned
// with one additional +1/+1 counter as it enters (a CR 614.1c "enters with").
func revelationsOfEzioReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(id, ctx.Controller(), false, map[string]int{game.CounterPlusOne: 1})
	return effects.NothingIfGone(err)
}
