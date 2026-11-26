package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sangrite Surge — Sorcery {4}{R}{G} (Shards of Alara, S82 OLD batch 511,
// #5588):
//
//	"Target creature gets +3/+3 and gains double strike until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d8053319-b2a0-4df3-9787-3591d12cf1b9",
		Name:         "Sangrite Surge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "double strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			if err := (effects.BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: "Sangrite Surge — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: "Sangrite Surge — double strike"}.Apply(ctx)
		},
	})
}
