package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lagrella, the Magpie — Legendary Creature — Human Soldier {G}{W}{U}, 2/3
// (Streets of New Capenna):
//
//	"When Lagrella enters, exile any number of other target creatures
//	 controlled by different players until Lagrella leaves the
//	 battlefield. When an exiled card enters under your control this
//	 way, put two +1/+1 counters on it."
//
// The targets are any number of other creatures with the set rule that no
// two share a controller (EachDifferentController, judged again on
// resolution, CR 608.2b). Each one still legal is exiled through
// ExileUntil{ThisLeaves}, so every card comes back the moment Lagrella
// leaves (CR 610.3), under its owner's control (CR 610.3c), and a
// Lagrella that left before the trigger resolved exiles nothing (CR
// 610.3b).
//
// The counters are a triggered ability, not "enters with" (#5877): each
// exile's return record carries it as ExileUntil.ThenEnters, and the
// return schedules it as a delayed trigger just before the card enters.
// It triggers only for the object the return made, and only when it
// enters under the control of Lagrella's controller — a card its owner
// gets back, when that owner is another player, gets none. A token has
// ceased to exist in exile and never comes back.
//
// No simplification.
func init() {
	lagrellaExiledOne = effects.CardTail("then/lagrella-the-magpie-exiled-one", func(ctx *effects.Context, o game.TailOutcome) error {
		return lagrellaExileNext(ctx, o.Tail.Carry)
	})
	effects.Register(effects.Spec{
		OracleID:     "6a98b586-5187-48a1-bde3-f6141f3e9cb1",
		Name:         "Lagrella, the Magpie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Lagrella, the Magpie — exile other creatures controlled by different players until Lagrella leaves", lagrellaExileEach),
				effects.Another(effects.TargetCreature("any number of other target creatures controlled by different players")).
					WithCount(0, 0).EachDifferent(effects.EachDifferentController()),
			),
		},
	})
}

// lagrellaExiledOne is one exile settled, then the next pick. An on-disk
// key, registered in init (the body names itself again).
var lagrellaExiledOne game.TailRef

// lagrellaPutTwoCountersOnIt is the follow-up, "put two +1/+1 counters
// on it": the returned permanent, if it is still the object the return
// made (Params.Object, CR 400.7). An on-disk key.
var lagrellaPutTwoCountersOnIt = game.DelayedBody("until-then/lagrella-two-plus-one-counters-on-it",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		obj := p.Object
		c, ok := effects.BattlefieldCardForEffect(g, obj.ID)
		if !ok || c.ObjectEpoch != obj.Epoch {
			return nil
		}
		return effects.AddCounter{Target: obj.ID, Kind: game.CounterPlusOne, N: 2}.Apply(effects.NewContext(g, item))
	})

// lagrellaExileEach is the enters trigger's effect: every pick still legal
// as it resolves.
func lagrellaExileEach(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var picks []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			picks = append(picks, t.ID)
		}
	}
	return lagrellaExileNext(ctx, picks)
}

// lagrellaExileNext exiles the first of `picks` until Lagrella leaves the
// battlefield, carrying the follow-up, the rest riding the exile's
// continuation as its Carry.
func lagrellaExileNext(ctx *effects.Context, picks []uuid.UUID) error {
	if len(picks) == 0 {
		return nil
	}
	id, rest := picks[0], picks[1:]
	t := ctx.ItemTail(lagrellaExiledOne)
	t.Carry = append([]uuid.UUID(nil), rest...)
	return effects.ExileUntil{
		Target:          id,
		ThisLeaves:      true,
		Label:           "Lagrella, the Magpie — returns when Lagrella leaves the battlefield",
		ThenTail:        t,
		ThenEnters:      lagrellaPutTwoCountersOnIt,
		ThenEntersLabel: "Lagrella, the Magpie — put two +1/+1 counters on the returned creature",
	}.Apply(ctx)
}
