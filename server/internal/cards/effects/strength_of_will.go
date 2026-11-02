package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Strength of Will — Instant {1}{G} (Marvel's Spider-Man, #4460):
//
//	"Until end of turn, target creature you control gains indestructible and
//	 "Whenever this creature is dealt damage, put that many +1/+1 counters
//	 on it.""
//
// Not Dead After All's duration grant: the target gains the granted trigger
// until end of turn (ADR 0093), together with indestructible. The trigger is
// Mockingbird's enrage over a batch of damage (CR 120.2) with the batch's
// total as the counter count ("that many", ThatMuchDamage); the counters go
// on only while the creature is still the same object on the battlefield.
// Indestructible keeps it there through lethal damage, so it grows instead of
// dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0389b48a-e64c-44fb-8a41-fd571cba237c",
		Name:         "Strength of Will",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive | game.AnswerPump},
		Targets:      TargetCreature("target creature you control", YouControl()),
		Grants: []AbilityGrant{{
			Key: strengthOfWillGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source)
				},
				Key:        "Strength of Will — put that many +1/+1 counters on this creature",
				AtBatchEnd: ThatMuchDamage(),
				Effect:     strengthOfWillCounters,
			}},
			Text: "Whenever this creature is dealt damage, put that many +1/+1 counters on it.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"indestructible"},
				Label:    "Strength of Will — indestructible until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{strengthOfWillGrant},
				Label:  "Strength of Will — grows when dealt damage",
			}.Apply(ctx)
		},
	})
}

const strengthOfWillGrant = "strength-of-will/damage-counters"

// strengthOfWillCounters is "put that many +1/+1 counters on it": the
// batch's damage total, on the creature that was dealt it while it is still
// that object on the battlefield.
func strengthOfWillCounters(g *game.Game, item *game.StackItem) error {
	if !sourceIsStillThisPermanent(g, item) {
		return nil
	}
	n := NewContext(g, item).Trigger().Event.Amount
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
}
