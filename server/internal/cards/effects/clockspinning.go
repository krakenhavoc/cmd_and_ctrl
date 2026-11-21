package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Clockspinning — Instant {U} (Doctor Who):
//
//	"Buyback {3} (You may pay an additional {3} as you cast this spell. If
//	 you do, put this card into your hand as it resolves.)
//	 Choose a counter on target permanent or suspended card. Remove that
//	 counter from that permanent or card or put another of those counters on
//	 it."
//
// The target is a permanent or a suspended card in exile (a card with suspend
// and a time counter on it, CR 702.62b). The choice of counter and of "remove
// one" or "put another" is one prompt: one option per kind of counter on it
// and per action (the counters of time travel, CR 701.56a, are handled the
// same way). A target with no counter has nothing to choose. Buyback is the
// shared additional cost (CR 702.27a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "3e1b32a9-27a4-4550-a644-ebe254e1f512",
		Name:          "Clockspinning",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Buyback("{3}")},
		Targets:       clockspinningTarget(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return clockspinningAsk(ctx, t.ID)
			}
			return nil
		},
	})
}

// clockspinningTarget is "target permanent or suspended card": a permanent,
// or a card in exile that has suspend and a time counter on it. Mode "any"
// (and no players) lets the client's picker reach the battlefield and exile.
func clockspinningTarget() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "any",
		Label: "target permanent or suspended card",
		Zones: []game.ZoneKind{game.ZoneBattlefield, game.ZoneExile},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, zone game.ZoneKind) bool {
			if zone == game.ZoneBattlefield {
				return true
			}
			return zone == game.ZoneExile && game.CardHasSuspend(c) && c.Counters[game.CounterTime] > 0
		},
		Min: 1, Max: 1,
	}
}

func clockspinningAsk(ctx *Context, id uuid.UUID) error {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	kinds := gapB94CounterKinds(c)
	if len(kinds) == 0 {
		return nil
	}
	var options []game.ChoiceOption
	for _, kind := range kinds {
		options = append(options,
			game.ChoiceOption{Label: "Remove a " + kind + " counter from " + c.Name},
			game.ChoiceOption{Label: "Put another " + kind + " counter on " + c.Name})
	}
	return PickOption{
		Question: "Clockspinning — choose a counter on " + c.Name + ", then remove it or put another of that kind on it",
		Options:  options,
		ThenKey:  clockspinningAnswered,
		Carry:    []uuid.UUID{id},
		WithItem: true,
	}.Apply(ctx)
}

// clockspinningAnswered applies the pick: option 2k removes a counter of the
// k-th kind (in sorted order) and 2k+1 adds one. An on-disk key.
var clockspinningAnswered = OptionPickThen("option-pick/clockspinning-counter", func(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) != 1 || r.Index < 0 {
		return nil
	}
	id := r.Carry[0]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	kinds := gapB94CounterKinds(c)
	k := r.Index / 2
	if k >= len(kinds) {
		return nil
	}
	delta := -1
	if r.Index%2 == 1 {
		delta = 1
	}
	return ctx.Game.AddCounterByForEffect(r.Chooser, id, kinds[k], delta)
})
