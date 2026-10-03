package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magma Pummeler — Creature — Elemental {X}{R}{R}, 0/0:
//
//	"This creature enters with X +1/+1 counters on it.
//	 If damage would be dealt to this creature while it has a +1/+1 counter on it, prevent that damage and remove that many +1/+1 counters from it. When one or more counters are removed from this creature this way, it deals that much damage to any target."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect removes
// "that many" counters — the damage, prevented or not (CR 615.12) — or
// every counter it has when that is fewer. "When one or more counters are
// removed this way" is a reflexive trigger (CR 603.12), put on the stack
// with that damage as the amount: dealt more damage than it has counters,
// all of it is prevented, every counter comes off, it usually dies, and
// it still deals that much damage to any target, as it last existed (the
// ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "1a7c5807-afdc-4855-afd5-39839d96fc77",
		Name:                       "Magma Pummeler",
		Completeness:               CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				While: WhileItHasAPlusOneCounter,
				Then:  magmaPummelerRemoveBody,
				Label: "Magma Pummeler — prevent damage to it and remove that many +1/+1 counters",
			}),
		},
	})
}

var (
	// The additional effect: remove the counters, then the reflexive
	// trigger when at least one came off.
	magmaPummelerRemoveBody = game.DelayedBody("magma-pummeler/remove-counters", magmaPummelerRemove)
	// The reflexive trigger: that much damage to any target.
	magmaPummelerStrikeBody = game.ReflexiveBody("magma-pummeler/strike", breechesBlastEffect, constTargets(TargetAny))
)

func magmaPummelerRemove(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	info, ok := followUpThis(g, item)
	if !ok {
		return nil
	}
	n := min(thatDamage(item), info.Counters[game.CounterPlusOne])
	if n <= 0 {
		return nil
	}
	if err := g.AddCounterForEffect(item.SourceCardID, game.CounterPlusOne, -n); err != nil {
		return err
	}
	t := WhenYouDo("Magma Pummeler — deal that much damage to any target", magmaPummelerStrikeBody)
	t.Params = game.EffectParams{Amount: thatDamage(item)}
	return t.Apply(NewContext(g, item))
}
