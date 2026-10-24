package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Strong, the Brutish Thespian — Legendary Creature — Mutant Berserker
// {4}{G}{G}, 7/7:
//
//	"Ward {2}
//	 Enrage — Whenever Strong is dealt damage, you get three rad counters
//	 and put three +1/+1 counters on Strong.
//	 You gain life rather than lose life from radiation."
//
// #2809. Enrage triggers once per damage EVENT, not once per source:
// damage dealt at the same time is one event (CR 120.3, CR 510.2), so
// Strong blocked by three creatures triggers once. The engine emits one
// EventDealDamage per source, so the ability waits for the damage to
// finish (AtBatchEnd, ThatMuchDamage, Ill-Tempered Loner's shape, ADR
// 0108 amendment 2026-10-13) and fires once per damage instance: one
// combat damage step, or one damage instruction. Two instructions in one
// resolution are two events and two triggers, as printed. It fires
// while Strong is still on the battlefield, so lethal damage still
// triggers it; the counters then have nowhere to go (CR 608.2b), and the
// rad counters still arrive.
//
// The last line is a CR 614 replacement on the life change the rad
// counters' own trigger makes (CR 728.1a, ReplacementEvent.
// LifeFromRadiation): a loss from radiation by Strong's controller
// becomes a gain of the same amount. Any other life loss is untouched,
// and the trigger still removes one rad counter per nonland card milled.
//
// No simplification.
func init() {
	const enrage = "Strong, the Brutish Thespian — Enrage: you get three rad counters and put three +1/+1 counters on Strong"
	Register(Spec{
		OracleID:     "f0ef48e1-c3d7-4ce2-8f7c-653342778cae",
		Name:         "Strong, the Brutish Thespian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Strong, the Brutish Thespian — ward {2}"),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b35SelfWasDealtDamage(ev, source)
				},
				Key:        enrage,
				AtBatchEnd: ThatMuchDamage(),
				Effect: Do(
					youGetRadCounters{N: 3},
					CounterOnThis{Kind: game.CounterPlusOne, N: 3},
				),
			},
		},
		Replacements: []game.ReplacementEffect{youGainLifeRatherThanLoseLifeFromRadiation(
			"Strong, the Brutish Thespian — you gain life rather than lose life from radiation")},
	})
}

// youGainLifeRatherThanLoseLifeFromRadiation is "You gain life rather
// than lose life from radiation" (CR 728.1a): a replacement on a life
// LOSS marked LifeFromRadiation for the source's controller, which
// turns it into a gain of the same amount.
func youGainLifeRatherThanLoseLifeFromRadiation(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventLife && ev.LifeFromRadiation && ev.LifeDelta < 0 &&
				src != nil && src.Controller != uuid.Nil && ev.LifePlayer == src.Controller
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.LifeDelta = -ev.LifeDelta
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}
