package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mukotai Soulripper — Artifact — Vehicle {1}{B}, 4/3 (Kamigawa: Neon
// Dynasty, S82):
//
//	"Whenever this Vehicle attacks, you may sacrifice another artifact or
//	 creature. If you do, put a +1/+1 counter on this Vehicle and it gains
//	 menace until end of turn.
//	 Crew 2"
//
// Comet Crawler's attack trigger (maySacrificeOne, asked at resolution over
// what you control then, the source excluded). The counter and the menace go
// on only while the Vehicle is still the object that attacked (CR 400.7).
// Crew 2 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "060d109f-2969-4f9e-a723-535d662a571a",
		Name:         "Mukotai Soulripper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Mukotai Soulripper — you may sacrifice another artifact or creature for a +1/+1 counter and menace",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Mukotai Soulripper — sacrifice another artifact or creature to put a +1/+1 counter on it and give it menace until end of turn?",
						effects.Or(effects.Artifact(), effects.Creature()), true, mukotaiSoulripperSacrificed)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mukotai Soulripper"),
		}},
	})
}

var mukotaiSoulripperSacrificed = effects.RegisterSacrificedOne("prompt-run/mukotai-soulripper-grow", func(ctx *effects.Context, _ uuid.UUID) error {
	if effects.SourceIsNewObject(ctx.Game, ctx.Item) || !effects.OnBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"menace"}, Label: "Mukotai Soulripper — menace"}.Apply(ctx)
})
