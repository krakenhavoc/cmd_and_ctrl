package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blast Zone — Land:
//
//	"This land enters with a charge counter on it.
//	 {T}: Add {C}.
//	 {X}{X}, {T}: Put X charge counters on this land.
//	 {3}, {T}, Sacrifice this land: Destroy each nonland permanent with
//	 mana value equal to the number of charge counters on this land."
//
// Engineered Explosives' sweep on a land. "{X}{X}" is two X slots, so
// X=3 costs six mana and banks three counters, and X=0 is a legal,
// pointless announcement. The sacrifice is a cost, so the number is
// the land's last-known information (CR 608.2h), read through
// ctx.SourcePermanent rather than the graveyard card, whose counters
// are cleared. The sweep spares lands, including this one; a token's
// mana value is 0, so a Blast Zone with no counters destroys every
// nonland token, and an indestructible permanent survives it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "393a254f-be31-431a-9341-a51286f8cbce",
		Name:         "Blast Zone",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(game.CounterCharge, 1, "Blast Zone enters with a charge counter on it")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{X}{X}, {T}: Put X charge counters on this land.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{X}{X}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterCharge, N: ctx.X()}.Apply(ctx)
				},
			},
			{
				Label:   "{3}, {T}, Sacrifice this land: Destroy each nonland permanent with mana value equal to the number of charge counters on this land.",
				Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.DestroyNonlandsWithManaValueOfChargeCounters,
			},
		},
	})
}
