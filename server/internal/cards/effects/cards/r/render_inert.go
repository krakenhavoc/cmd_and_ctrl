package cardsr

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Render Inert — Sorcery {2}{B} (March of the Machine):
//
//	"Remove up to five counters from target permanent.
//	 Draw a card."
//
// Glissa Sunslayer's counter choice (a registered counter-choice clause of
// its own, renderInertRemoves, with the same bot reading): as the spell resolves its controller chooses how many
// of each kind on the permanent to remove, up to five in all, whoever
// controls it (CR 608.2c). The draw does not depend on removing any; a
// permanent that is an illegal target as the spell resolves leaves the spell
// countered, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ece3c6-81c2-441a-acce-c109aaa9026b",
		Name:         "Render Inert",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				// The draw follows the removal once every counter question
				// has been answered, so it is the clause's final step.
				return renderInertRemoves.Choose(ctx, t.ID, nil, 5, "Render Inert — remove up to five counters")
			}
			return nil
		},
	})
}

// renderInertRemoves removes the counts chosen, then draws the card: the
// bot removes -1/-1 counters from its own permanent and everything else
// from another's, as Glissa's clause does.
var renderInertRemoves = effects.RegisterCounterChoice("number/render-inert-remove-counters",
	func(ctx *effects.Context, target uuid.UUID, kind string, most int) int {
		mine := false
		if c, ok := ctx.Game.LookupCardForEffect(target); ok {
			mine = c.Controller == ctx.Controller()
		}
		if (kind == "-1/-1") == mine {
			return most
		}
		return 0
	},
	func(ctx *effects.Context, target uuid.UUID, _ []uuid.UUID, picked map[string]int) error {
		kinds := make([]string, 0, len(picked))
		for kind := range picked {
			kinds = append(kinds, kind)
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			if err := (effects.AddCounter{Target: target, Kind: kind, N: -picked[kind]}).Apply(ctx); err != nil {
				return err
			}
		}
		return effects.DrawCards{N: 1}.Apply(ctx)
	})
