package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freyalise's Winds — Enchantment {2}{G}{G} (Ice Age, S82 OLD batch 845, #5636):
//
//	"Whenever a permanent becomes tapped, put a wind counter on it.
//	 If a permanent with a wind counter on it would untap during its
//	 controller's untap step, remove all wind counters from it instead."
//
// The trigger is Magda's "became tapped" reading, for any permanent whoever
// controls it: an EventTapCard, or an EventAttack of an attacker that is now
// tapped (the engine taps an attacker without an EventTapCard, and a vigilance
// attacker did not become tapped). The counter goes on the permanent the
// event named, read as the trigger resolves (CR 608.2h), if it is still there.
//
// The replacement is an untap-step restriction that also clears the counters
// (game.UntapStepRestriction.ClearsCounter): the permanent stays tapped
// through its controller's untap step and loses every wind counter instead.
// It reads the counters live, so a permanent keeps them, and untaps as usual,
// once the Winds are gone; an untap by an effect is not replaced (CR 502.3).
//
// No simplification.
func init() {
	const wind = "wind"
	effects.Register(effects.Spec{
		OracleID:     "ffeb792a-6df0-46c1-badc-4deb857f7d48",
		Name:         "Freyalise's Winds",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return false
					}
					return ev.Kind == game.EventTapCard || c.Tapped
				},
				"Freyalise's Winds — put a wind counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					obj := ctx.Trigger().Object
					if obj == nil {
						return nil
					}
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					return effects.AddCounter{Target: obj.Ref().ID, Kind: wind, N: 1}.Apply(ctx)
				}),
		},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label:         "Freyalise's Winds — a permanent with a wind counter loses its wind counters instead of untapping",
			ClearsCounter: wind,
			Restricts: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.Counters[wind] > 0
			},
		}},
	})
}
