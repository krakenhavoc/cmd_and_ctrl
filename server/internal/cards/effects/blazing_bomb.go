package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Bomb — Creature — Elemental {R}, 1/1 (Final Fantasy, #4462):
//
//	"Whenever you cast a noncreature spell, if at least four mana was
//	 spent to cast it, put a +1/+1 counter on this creature.
//	 Blow Up — {T}, Sacrifice this creature: It deals damage equal to its
//	 power to target creature. Activate only as a sorcery."
//
// The cast trigger reads the amount spent on the spell off its payment
// as the spell is cast (mana of every kind counts, generic paid with
// colored mana included). The intervening if (CR 603.4) is asked then;
// what was spent to cast a spell never changes, so asking again as the
// trigger resolves gives the same answer. The counter goes on only while
// the Bomb is still on the battlefield.
//
// Blow Up is Hellhole Flailer's shape: the sacrifice is part of the
// cost, so "its power" is the Bomb's power as it last existed on the
// battlefield (CR 113.7a), counters included, and it is still the
// damage's source. "Blow Up" is a flavor word with no rules meaning.
//
// No simplification.
func init() {
	grow := On(game.EventCast, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return YouCast(Noncreature())(ev, source, lki, g) && g.StackItemPaidForEffect(ev.CardID).Spent().Total() >= 4
	}, "Blazing Bomb — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))
	Register(Spec{
		OracleID:       "088acd15-029b-4091-960a-d5060a364ca9",
		Name:           "Blazing Bomb",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsYourCasts,
		Triggered:      []game.TriggeredAbility{grow},
		Activated: []ActivatedAbility{{
			Label:        "Blow Up — {T}, Sacrifice this creature: It deals damage equal to its power to target creature. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), SacrificeThis()),
			Targets:      TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				target, legal := b16FirstLegalTargetCard(ctx)
				if !ok || !legal || info.Power <= 0 {
					return nil
				}
				return DealDamage{Source: item.SourceCardID, Target: target, Amount: info.Power}.Apply(ctx)
			},
		}},
	})
}
