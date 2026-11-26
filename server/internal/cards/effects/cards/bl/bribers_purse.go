package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Briber's Purse — Khans of Tarkir (S82 OLD b346, #5554):
//
//	Artifact {X}: This artifact enters with X gem counters on it.
//	{1}, {T}, Remove a gem counter from this artifact: Target creature can't attack
//	or block this turn.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb2cc0f4-6abd-48d6-b442-76eafbfd2309",
		Name:         "Briber's Purse",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.XCounters(effects.OldB346GemCounter),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Remove a gem counter from this artifact: Target creature can't attack or block this turn.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.RemoveCountersFromThis(effects.OldB346GemCounter, 1)),
			Targets: effects.TargetCreature("target creature"),
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantAttackOrBlock, Label: "Briber's Purse — can't attack or block"}.Apply(ctx)
			},
		}},
	})
}
