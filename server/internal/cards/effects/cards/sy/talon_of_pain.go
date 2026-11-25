package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talon of Pain — Artifact {4} (S82 GAP2 batch 04, #5948):
//
//	"Whenever a source you control other than this artifact deals damage
//	 to an opponent, put a charge counter on this artifact.
//	 {X}, {T}, Remove X charge counters from this artifact: It deals X
//	 damage to any target."
//
// The counter trigger is Quest for Pure Flame's damage watcher, one
// counter per damage event, without the "may" and without the Talon
// itself as a source. The activation announces ONE X for the {X} and the
// counters removed (CR 107.3a, #6085's RemoveCountersXFromThisSameX), so
// X is bounded by the charge counters the Talon holds, and the damage is
// that X. The artifact stays on the battlefield and is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0e4d2fbc-1b34-4501-9022-8ea800cb2d24",
		Name:         "Talon of Pain",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Amount <= 0 || ev.Source == source.InstanceID || ev.Target == source.Controller || g.PlayerByIDForEffect(ev.Target) == nil {
					return false
				}
				controller, ok := effects.B34DamageSourceController(ev, g)
				return ok && controller == source.Controller
			}, "Talon of Pain — put a charge counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}, {T}, Remove X charge counters from this artifact: It deals X damage to any target.",
			Purpose: effects.ForTargets(effects.DamageXToTarget(0)),
			Cost:    effects.Plus(effects.ManaCost("{X}"), effects.TapCost(), effects.RemoveCountersXFromThisSameX(game.CounterCharge)),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: ctx.X()}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
