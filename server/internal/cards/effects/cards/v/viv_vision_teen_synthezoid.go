package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viv Vision, Teen Synthezoid — {3} Legendary Artifact Creature — Robot
// Hero 2/2 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Cybernetic Senses — Whenever Viv Vision attacks, draw a card if her
//	 power is 4 or greater.
//	 Power-up — {7}: Put two +1/+1 counters on Viv Vision. (Activate each
//	 power-up ability only once. Reduce the cost by her mana cost if she
//	 entered this turn.)"
//
// The power is read as the trigger resolves, from the layered value (counters
// and pumps put on her in response count); a Viv Vision who has left reads
// her last-known power (CR 608.2h). Power-up is the engine's (CR 702.193).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f67a599c-200a-4831-9e73-126c7435a396",
		Name:            "Viv Vision, Teen Synthezoid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.Self, "Viv Vision, Teen Synthezoid — draw a card if her power is 4 or greater",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "Viv Vision's power is 4 or greater", Test: func(ctx *effects.Context) bool {
						return effects.PowerNowOrLastKnown(ctx.Game, ctx.Source()) >= 4
					}},
					Then: []effects.Applier{effects.DrawCards{N: 1}},
				})),
		},
		Activated: []effects.ActivatedAbility{
			effects.PowerUpAnswering("{7}: Put two +1/+1 counters on Viv Vision.",
				effects.ManaCost("{7}"), effects.PlusOneCountersOnThis(2)),
		},
	})
}
