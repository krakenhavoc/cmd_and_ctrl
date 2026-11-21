package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eldrazi Linebreaker — Creature — Eldrazi {1}{C}{R}, 3/3 (Modern Horizons 3):
//
//	"Devoid
//	 Trample
//	 At the beginning of combat on your turn, target creature you control gains
//	 haste and gets +X/+0 until end of turn, where X is the number of Eldrazi you
//	 control."
//
// X is counted as the ability resolves.
func init() {
	const label = "Eldrazi Linebreaker — target creature you control gains haste and gets +X/+0 until end of turn, where X is the number of Eldrazi you control"
	trigger := effects.Targeting(effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := len(effects.PermanentsControlledByMatching(g, item.Controller, effects.Subtype("Eldrazi")))
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{Target: id, Power: x, Label: label}.Apply(ctx)
		}), effects.TargetCreature("target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "db432387-58d3-4473-9442-b5b586dbea38",
		Name:            "Eldrazi Linebreaker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
