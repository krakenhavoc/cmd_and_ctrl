package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monoist Circuit-Feeder — Artifact Creature — Nautilus {4}{B}{B}, 4/4
// (Edge of Eternities, #4461):
//
//	"Flying
//	 When this creature enters, until end of turn, target creature you
//	 control gets +X/+0 and target creature an opponent controls gets -0/-X,
//	 where X is the number of artifacts you control."
//
// Two target clauses on one enters trigger (Assert Perfection's shape),
// each judged on its own as the trigger goes on the stack and again as it
// resolves (CR 608.2b): a clause whose creature has gone gets nothing, and
// the other still does. X is counted once as the ability resolves, through
// the layers, so the Circuit-Feeder itself counts, and is the same number
// for both halves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de70a6bb-32ed-4a89-a522-c2a71f08fb02",
		Name:            "Monoist Circuit-Feeder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Monoist Circuit-Feeder — target creature you control gets +X/+0 and target creature an opponent controls gets -0/-X",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsArtifact() })
					if x <= 0 {
						return nil
					}
					if mine, ok := ctx.ClauseTarget(0); ok && mine.Kind == game.TargetCard {
						if err := (effects.BoostUntilEOT{Target: mine.ID, Power: x, Label: "Monoist Circuit-Feeder — +X/+0"}).Apply(ctx); err != nil {
							return err
						}
					}
					if theirs, ok := ctx.ClauseTarget(1); ok && theirs.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: theirs.ID, Toughness: -x, Label: "Monoist Circuit-Feeder — -0/-X"}.Apply(ctx)
					}
					return nil
				}),
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				)),
		},
	})
}
