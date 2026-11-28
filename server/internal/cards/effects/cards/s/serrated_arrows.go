package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serrated Arrows — Artifact {4}, (Homelands, S82 OLD b822, #5635):
//
//	"This artifact enters with three arrowhead counters on it.
//	At the beginning of your upkeep, if there are no arrowhead counters on this
//	artifact, sacrifice it.
//	{T}, Remove an arrowhead counter from this artifact: Put a -1/-1 counter on
//	target creature."
//
// Entering counters (B10EntersWithCounters), an upkeep trigger with an
// intervening "if" (CR 603.4) judged as it triggers and again as it resolves,
// and a counter removed as a cost (paid at announce, so a response cannot spend
// it twice).
//
// No simplification.
func init() {
	const arrowhead = "arrowhead"
	none := func(g *game.Game, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && c.Counters[arrowhead] == 0
	}
	upkeep := effects.AtYourUpkeep("Serrated Arrows — sacrifice it", effects.SacrificeThisIfStillOnBattlefield)
	when := upkeep.AppliesTo
	upkeep.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return when(ev, source, lki, g) && source.Counters[arrowhead] == 0
	}
	effects.Register(effects.Spec{
		OracleID:     "5fa4607a-6731-4599-80f3-19d6097b4b28",
		Name:         "Serrated Arrows",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(arrowhead, 3, "Serrated Arrows: enters with three arrowhead counters on it"),
		},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "there are no arrowhead counters on this artifact", Test: func(ctx *effects.Context) bool {
				return none(ctx.Game, ctx.Source())
			}}, upkeep),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Remove an arrowhead counter from this artifact: Put a -1/-1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.RemoveCountersFromThis(arrowhead, 1)),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB822TargetGetsACounter(game.CounterMinusOne),
		}},
	})
}
