package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loki Laufeyson — Legendary Creature — God Sorcerer Villain {1}{R}, 2/1
// (Marvel Super Heroes, #4455):
//
//	"{1}, {T}: When you next cast an instant or sorcery spell with mana
//	 value less than or equal to Loki's power this turn, copy that spell.
//	 You may choose new targets for the copy.
//	 Power-up — {4}{R}: Put two +1/+1 counters on Loki. (Activate each
//	 power-up ability only once. Reduce the cost by his mana cost if he
//	 entered this turn.)"
//
// Doublecast's delayed trigger (WhenYouNextCast shape: it waits for a cast,
// a cast that does not match leaves it waiting, it ends with the turn) with
// one more test on the cast, loki's power. The power is read when the spell
// is cast: Loki's live power while he is on the battlefield, so counters
// put on him in the meantime count; once he has left, the power he had when
// the ability resolved (CR 608.2h's last-known information, recorded in the
// condition's params). The copy is the shared copy-the-spell body, which
// may choose new targets (CR 707.10c). Power-up is the engine's
// (CR 702.193).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c78778a-6335-4949-8ade-11d0f085cb2b",
		Name:         "Loki Laufeyson",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: When you next cast an instant or sorcery spell with mana value less than or equal to Loki's power this turn, copy that spell. You may choose new targets for the copy.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  lokiSetUpCopy,
			},
			lokiPowerUp(),
		},
	})
}

// lokiPowerUp is "Power-up — {4}{R}: Put two +1/+1 counters on Loki."
func lokiPowerUp() ActivatedAbility {
	ab := PowerUp("{4}{R}: Put two +1/+1 counters on Loki.", ManaCost("{4}{R}"), plusOneCountersOnThis(2))
	ab.Purpose.Answers = game.AnswerPump
	return ab
}

// lokiNextCastCondition is the delayed trigger's condition: a cast by the
// trigger's controller of an instant or sorcery whose mana value is at most
// Loki's power.
var lokiNextCastCondition = game.DelayedCondition("loki-laufeyson/next-cast", func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if !youNextCast(ev, dt, g, p) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	power := p.Amount
	if onBattlefield(g, dt.SourceCardID) {
		power = b31CurrentPowerOf(g, dt.SourceCardID)
	}
	return spell.ManaValue() <= power
})

// lokiSetUpCopy is the ability's resolution: schedule the delayed trigger,
// recording Loki's power now as the fallback for a Loki who has left.
func lokiSetUpCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	power := powerNowOrLastKnown(g, item.SourceCardID)
	return DelayedOnEvent{
		Label:      "Loki Laufeyson — copy that spell",
		On:         []game.EventKind{game.EventCast},
		Condition:  lokiNextCastCondition,
		CondParams: game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}, Amount: power},
		Body:       copyTheSpellBody,
	}.Apply(ctx)
}
