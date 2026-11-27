package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inner-Flame Igniter — Creature — Elemental Warrior {2}{R}, 2/2 (Lorwyn,
// S82 OLD b545, #5593):
//
//	"{2}{R}: Creatures you control get +1/+0 until end of turn. If this is
//	 the third time this ability has resolved this turn, creatures you
//	 control gain first strike until end of turn."
//
// The count is the per-turn tally of this Igniter's resolutions of this
// ability (Game.ResolvedThisTurn), the resolving one included, per object
// (CR 400.7). Only the third resolution grants first strike; the creatures
// are picked as the ability resolves (CR 611.2c). A resolution after the
// Igniter has left still pumps, since the ability is independent of its
// source (CR 113.7a), but the count is the old object's.
//
// No simplification.
func init() {
	const label = "{2}{R}: Creatures you control get +1/+0 until end of turn. If this is the third time this ability has resolved this turn, creatures you control gain first strike until end of turn."
	effects.Register(effects.Spec{
		OracleID:     "059788eb-2c55-4da2-bf54-aa2fa7954169",
		Name:         "Inner-Flame Igniter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   label,
			Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				mine := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Label: "Inner-Flame Igniter — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if g.ResolvedThisTurn(item.SourceCardID, label) != 3 {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"first strike"}, Label: "Inner-Flame Igniter — first strike"}.Apply(ctx)
			},
		}},
	})
}
