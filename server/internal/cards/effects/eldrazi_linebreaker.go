package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			x := len(permanentsControlledByMatching(g, item.Controller, Subtype("Eldrazi")))
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: label}).Apply(ctx); err != nil {
				return err
			}
			return BoostUntilEOT{Target: id, Power: x, Label: label}.Apply(ctx)
		}), TargetCreature("target creature you control", YouControl()))
	Register(Spec{
		OracleID:        "db432387-58d3-4473-9442-b5b586dbea38",
		Name:            "Eldrazi Linebreaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "trample"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
